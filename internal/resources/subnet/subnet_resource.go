package subnet

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	tagshelper "github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing subnet resources.
type ResourceClient interface {
	CreateSubnet(ctx context.Context, vpcName, name, cidr, vpcID, subnetType string, tags []client.Tag) (*client.CreateSubnetResponse, error)
	ListSubnetsForVPC(ctx context.Context, vpcName string) ([]*client.Subnet, error)
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
	CIDR      types.String `tfsdk:"cidr"`
	Type      types.String `tfsdk:"type"`
	VPCID     types.String `tfsdk:"vpc_id"`
	VPCName   types.String `tfsdk:"vpc_name"`
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
		Attributes:  ResourceAttributes(),
	}
}

// ResourceAttributes return the subnet terraform schema attributes
func ResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique identifier for the subnet (uuid).",
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
		"vpc_name": schema.StringAttribute{
			Description: "The name of the parent VPC.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"vpc_id": schema.StringAttribute{
			Description: "The id of the parent VPC.",
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
			Description: "The last error encountered during subnet CRUD operations.",
			Computed:    true,
		},
		"tags": schema.MapAttribute{
			Description: "Customer-managed key/value tags.",
			Optional:    true,
			ElementType: types.StringType,
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

// Create creates a new subnet in the DSPC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := tagshelper.ToClient(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CreateSubnet(
		ctx,
		plan.VPCName.ValueString(),
		plan.Name.ValueString(),
		plan.CIDR.ValueString(),
		plan.VPCID.ValueString(),
		plan.Type.ValueString(),
		tags,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subnet",
			fmt.Sprintf("Could not create subnet: %s", err.Error()),
		)
		return
	}

	// The create response is a minimal acknowledgement, so fetch the full
	// subnet (in particular its status) via the list endpoint.
	created, err := r.findSubnet(ctx, plan.VPCName.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created subnet",
			fmt.Sprintf("Could not read created subnet: %s", err.Error()),
		)
		return
	}

	// client.Subnet has no VPCName field (the API only scopes subnets by VPC
	// name in the URL path), so the plan is updated in place rather than
	// rebuilt from the API response, to avoid losing plan.VPCName/plan.VPCID.
	plan.ID = types.StringValue(createSubnetStateID(plan.VPCName.ValueString(), created.Name))
	plan.URN = types.StringValue(created.URN)
	plan.Status = types.StringValue(created.Status)
	plan.LastError = types.StringValue(created.LastError)

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

	// state.ID is the source of truth for the VPC name: on import, only ID is
	// populated (state.VPCName is still null at this point).
	vpcName, subnetName, err := parseSubnetStateID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing subnet ID",
			err.Error(),
		)
		return
	}

	found, err := r.findSubnet(ctx, vpcName, subnetName)
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

	state.ID = types.StringValue(createSubnetStateID(vpcName, found.Name))
	state.Name = types.StringValue(found.Name)
	state.CIDR = types.StringValue(found.CIDR)
	state.Type = types.StringValue(found.Type)
	state.VPCName = types.StringValue(vpcName)
	state.URN = types.StringValue(found.URN)
	state.Status = types.StringValue(found.Status)
	state.LastError = types.StringValue(found.LastError)

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

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// findSubnet searches for a subnet by name in the list of subnets for a VPC.
func (r *Resource) findSubnet(ctx context.Context, vpcName, subnetName string) (*client.Subnet, error) {
	subnets, err := r.client.ListSubnetsForVPC(ctx, vpcName)
	if err != nil {
		return nil, err
	}

	for _, s := range subnets {
		if s.Name == subnetName {
			return s, nil
		}
	}

	return nil, fmt.Errorf("subnet %q not found in VPC %q", subnetName, vpcName)
}

// createSubnetStateID builds the composite resource ID used since there is no
// GET /subnets/{name} endpoint to look up a subnet by a single opaque ID.
func createSubnetStateID(vpcName, subnetName string) string {
	return vpcName + ":" + subnetName
}

// parseSubnetStateID splits a composite resource ID back into its VPC name
// and subnet name parts.
func parseSubnetStateID(id string) (vpcName, subnetName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid subnet id %q: expected format \"<vpc_name>:<subnet_name>\"", id)
	}
	return parts[0], parts[1], nil
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not found") ||
		len(err.Error()) > 14 && err.Error()[:14] == "API error 404:")
}

// ToClient converts the Terraform subnet models into client subnets for the API request.
func ToClient(subnets []ResourceModel) []client.Subnet {
	if subnets == nil {
		return nil
	}
	result := make([]client.Subnet, len(subnets))
	for i, s := range subnets {
		result[i] = client.Subnet{
			Name: s.Name.ValueString(),
			CIDR: s.CIDR.ValueString(),
			Type: s.Type.ValueString(),
		}
	}
	return result
}

// ToTerraform converts client subnets from the API response into Terraform subnet models.
func ToTerraform(ctx context.Context, subnets []client.Subnet, diags *diag.Diagnostics) []ResourceModel {
	if len(subnets) == 0 {
		return nil
	}
	result := make([]ResourceModel, len(subnets))
	for i, s := range subnets {
		result[i] = SubnetToTerraform(ctx, s, diags)
	}
	return result
}

// SubnetToTerraform converts client subnets from the API response into Terraform subnet models.
func SubnetToTerraform(ctx context.Context, subnet client.Subnet, diags *diag.Diagnostics) ResourceModel {
	return ResourceModel{
		ID:        types.StringValue(subnet.ID.String()),
		URN:       types.StringValue(subnet.URN),
		Name:      types.StringValue(subnet.Name),
		CIDR:      types.StringValue(subnet.CIDR),
		Type:      types.StringValue(subnet.Type),
		VPCID:     types.StringValue(subnet.VPCID.String()),
		Status:    types.StringValue(subnet.Status),
		LastError: types.StringValue(subnet.LastError),
		Tags:      tagshelper.ToTerraform(ctx, subnet.Tags, diags),
	}
}
