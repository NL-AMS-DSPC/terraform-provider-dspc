package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/subnet"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing VPC resources.
type ResourceClient interface {
	CreateVPC(ctx context.Context, id, name string, tags []client.Tag, subnets []client.CreateSubnetRequest) (client.VPC, error)
	GetVPC(ctx context.Context, name string) (client.VPC, error)
	DeleteVPC(ctx context.Context, name string) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID        types.String `tfsdk:"id"`
	URN       types.String `tfsdk:"urn"`
	Name      types.String `tfsdk:"name"`
	CIDR      types.String `tfsdk:"cidr"`
	Status    types.String `tfsdk:"status"`
	LastError types.String `tfsdk:"last_error" `
	Subnets   types.List   `tfsdk:"subnets"`
	Tags      types.Map    `tfsdk:"tags"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VPC in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the VPC.",
				Computed:    true,
			},
			"urn": schema.StringAttribute{
				Description: "The uniform resource name of the VPC.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the VPC. Must be unique within the namespace.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cidr": schema.StringAttribute{
				Description: "The CIDR range for the VPC (e.g. \"10.0.0.0/24\").",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the VPC (pending, active, deleting, error).",
				Computed:    true,
			},
			"last_error": schema.StringAttribute{
				Description: "The current status of the VPC (pending, active, deleting, error).",
				Computed:    true,
			},
			"subnets": schema.ListNestedAttribute{
				Description: "Subnets for this VPC.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: subnet.ResourceAttributes(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				Description: "User defined tags attached to the VPC.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Network
}

// Create creates a new VPC in the DSPC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpcTags := tags.ToClient(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	subnets := subnet.ToClient(ctx, plan.Subnets, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.CreateVPC(ctx, plan.ID.ValueString(), plan.Name.ValueString(), vpcTags, subnets)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VPC",
			fmt.Sprintf("Could not create VPC: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(vpc.ID)
	plan.URN = types.StringValue(vpc.URN)
	plan.CIDR = types.StringValue(vpc.CIDR)
	plan.Status = types.StringValue(vpc.Status)
	plan.LastError = types.StringValue(vpc.LastError)
	if !plan.Subnets.IsNull() {
		plan.Subnets = subnet.ToTerraformResourceList(ctx, vpc.Subnets, &resp.Diagnostics)
	}
	if !plan.Tags.IsNull() {
		plan.Tags = tags.ToTerraform(ctx, vpc.Tags, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vpc, err := r.client.GetVPC(ctx, state.Name.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting VPC",
			fmt.Sprintf("Could not get VPC: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(vpc.ID)
	state.URN = types.StringValue(vpc.URN)
	state.Name = types.StringValue(vpc.Name)
	state.CIDR = types.StringValue(vpc.CIDR)
	state.Status = types.StringValue(vpc.Status)
	state.LastError = types.StringValue(vpc.LastError)
	state.Subnets = subnet.ToTerraformResourceList(ctx, vpc.Subnets, &resp.Diagnostics)
	state.Tags = tags.ToTerraform(ctx, vpc.Tags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the VPC in the DSPC platform.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"VPC updates are not supported by the DSPC API. Changes require VPC recreation. "+
			"Consider using lifecycle { ignore_changes = [name] } if you need to prevent replacement.",
	)
}

// Delete deletes the VPC in the DSPC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVPC(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VPC",
			fmt.Sprintf("Could not delete VPC: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the VPC from the DSPC platform.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (err.Error() == "resource not found" ||
		len(err.Error()) > 14 && err.Error()[:14] == "API error 404:")
}
