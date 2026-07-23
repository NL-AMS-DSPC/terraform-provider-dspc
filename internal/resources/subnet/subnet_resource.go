package subnet

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}

	resourceAttrTypes = map[string]attr.Type{
		"id":         types.StringType,
		"urn":        types.StringType,
		"name":       types.StringType,
		"vpc_name":   types.StringType,
		"vpc_id":     types.StringType,
		"cidr":       types.StringType,
		"type":       types.StringType,
		"status":     types.StringType,
		"last_error": types.StringType,
		"tags":       types.MapType{ElemType: types.StringType},
	}
)

// ResourceClient defines the interface for managing subnet resources.
type ResourceClient interface {
	CreateSubnet(ctx context.Context, vpcName string, request client.CreateSubnetRequest) (client.Subnet, error)
	GetSubnet(ctx context.Context, vpcName, subnetName string) (client.Subnet, error)
	DeleteSubnet(ctx context.Context, vpcName, subnetName string) error
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
	VPCName   types.String `tfsdk:"vpc_name"`
	VPCID     types.String `tfsdk:"vpc_id"`
	CIDR      types.String `tfsdk:"cidr"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	LastError types.String `tfsdk:"last_error"`
	Tags      types.Map    `tfsdk:"tags"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a subnet within a VPC in the DSPC platform.",
		Attributes:  ResourceAttributes(true),
	}
}

// ResourceAttributes returns shema attributes for a subnet resource
func ResourceAttributes(requireVPCAttributes bool) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique identifier for the subnet.",
			Computed:    true,
		},
		"urn": schema.StringAttribute{
			Description: "The uniform resource name for the subnet.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of the subnet. Must be unique within the VPC.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"cidr": schema.StringAttribute{
			Description: "The CIDR range for the subnet (e.g. \"10.0.0.0/25\"). Must be within the VPC CIDR range.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"type": schema.StringAttribute{
			Description: "The type of the subnet: \"public\" or \"private\".",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"status": schema.StringAttribute{
			Description: "The current status of the subnet (pending, active, deleting, error).",
			Computed:    true,
		},
		"last_error": schema.StringAttribute{
			Description: "The last error encountered during CRUD of the subnet.",
			Computed:    true,
		},
		"tags": schema.MapAttribute{
			Description: "User defined tags attached to the subnet.",
			Optional:    true,
			ElementType: types.StringType,
		},
	}

	if requireVPCAttributes {
		attrs["vpc_name"] = schema.StringAttribute{
			Description: "The name of the parent VPC.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		}
		attrs["vpc_id"] = schema.StringAttribute{
			Description: "The id of the parent VPC.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		}
	} else {
		attrs["vpc_id"] = schema.StringAttribute{
			Description: "The id of the parent VPC.",
			Computed:    true,
		}
		attrs["vpc_name"] = schema.StringAttribute{
			Description: "The name of the parent VPC.",
			Computed:    true,
		}
	}

	return attrs
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

// Create creates a new subnet in the DSPC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subnet, err := r.client.CreateSubnet(
		ctx,
		plan.VPCName.ValueString(),
		client.CreateSubnetRequest{
			VPCID: plan.VPCID.ValueString(),
			Name:  plan.Name.ValueString(),
			CIDR:  plan.CIDR.ValueString(),
			Type:  plan.Type.ValueString(),
			Tags:  tags.ToClient(ctx, plan.Tags, &resp.Diagnostics),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet",
			fmt.Sprintf("Could not create subnet: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(subnet.ID)
	plan.URN = types.StringValue(subnet.URN)
	plan.Status = types.StringValue(subnet.Status)
	plan.LastError = types.StringValue(subnet.LastError)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
// Since there is no GET /subnets/{name} endpoint, we use ListSubnetsForVPC and find by name.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	subnet, err := r.client.GetSubnet(ctx, state.VPCName.ValueString(), state.Name.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting subnet",
			fmt.Sprintf("Could not get subnet: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(subnet.ID)
	state.URN = types.StringValue(subnet.URN)
	state.Name = types.StringValue(subnet.Name)
	state.CIDR = types.StringValue(subnet.CIDR)
	state.Type = types.StringValue(subnet.Type)
	state.VPCID = types.StringValue(subnet.VPCID)
	state.Status = types.StringValue(subnet.Status)
	state.LastError = types.StringValue(subnet.LastError)
	state.Tags = tags.ToTerraform(ctx, subnet.Tags, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the subnet in the DSPC platform.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Subnet updates are not supported by the DSPC API. Changes require subnet recreation. "+
			"Consider using lifecycle { ignore_changes = [name] } if you need to prevent replacement.",
	)
}

// Delete deletes the subnet in the DSPC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSubnet(ctx, state.VPCName.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subnet",
			fmt.Sprintf("Could not delete subnet: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the subnet from the DSPC platform.
// The import ID should be in the format: "vpc-name:subnet-name"
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format 'vpc-name:subnet-name', got: %s", req.ID),
		)
		return
	}

	vpcName := parts[0]
	subnetName := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vpc_name"), vpcName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), subnetName)...)
}

// createSubnetStateID creates a unique identifier for the subnet resource.
func createSubnetStateID(vpcName, subnetName string) string {
	return vpcName + ":" + subnetName
}

// splitImportID splits an import ID string by the first colon separator.
func splitImportID(id string) []string {
	idx := strings.Index(id, ":")
	if idx < 0 {
		return []string{id}
	}
	return []string{id[:idx], id[idx+1:]}
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not found") ||
		len(err.Error()) > 14 && err.Error()[:14] == "API error 404:")
}

// ToClient converts the Terraform subnets model into client subnets for the API request.
func ToClient(ctx context.Context, subnets types.List, diags *diag.Diagnostics) []client.CreateSubnetRequest {
	if subnets.IsNull() || subnets.IsUnknown() {
		return nil
	}

	var models []ResourceModel
	diags.Append(subnets.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	res := make([]client.CreateSubnetRequest, len(models))
	for i, m := range models {
		res[i] = client.CreateSubnetRequest{
			Name:  m.Name.ValueString(),
			CIDR:  m.CIDR.ValueString(),
			VPCID: m.VPCID.ValueString(),
			Type:  m.Type.ValueString(),
			Tags:  tags.ToClient(ctx, m.Tags, diags),
		}
	}

	return res
}

// ToTerraformResourceList converts a slice of client.Subnet into a slice of terraform List
func ToTerraformResourceList(ctx context.Context, subnets []client.Subnet, diags *diag.Diagnostics) types.List {
	if len(subnets) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: resourceAttrTypes})
	}

	models := make([]ResourceModel, len(subnets))
	for i, s := range subnets {
		models[i] = ToTerraformResource(ctx, s, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: resourceAttrTypes}, models)
	diags.Append(d...)

	return list
}

// ToTerraformResource converts a client.Subnet into a terraform ResourceModel
func ToTerraformResource(ctx context.Context, s client.Subnet, diags *diag.Diagnostics) ResourceModel {
	return ResourceModel{
		ID:        types.StringValue(s.ID),
		URN:       types.StringValue(s.URN),
		Name:      types.StringValue(s.Name),
		CIDR:      types.StringValue(s.CIDR),
		Type:      types.StringValue(s.Type),
		VPCID:     types.StringValue(s.VPCID),
		Status:    types.StringValue(s.Status),
		LastError: types.StringValue(s.LastError),
		Tags:      tags.ToTerraform(ctx, s.Tags, diags),
	}
}
