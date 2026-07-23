// Package role provides Terraform resources and data sources for managing roles.
package role

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing role resources.
type ResourceClient interface {
	CreateRole(ctx context.Context, name string, actions []string) error
	GetRole(ctx context.Context, name string) (*client.Role, error)
	UpdateRole(ctx context.Context, name string, actions []string) (*client.Role, error)
	DeleteRole(ctx context.Context, name string) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Actions types.List   `tfsdk:"actions"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a role in the authorization service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the role (equals the role name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the role. Must be unique within the tenant.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"actions": schema.ListAttribute{
				Description: "The list of permission actions assigned to this role (e.g. \"vm:CreateVM\", \"uam:ListUsers\").",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure stores the provider-configured client on the resource.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	ascClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if ascClient.Authorization == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected authorization service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = ascClient.Authorization
}

// Create creates a new role in the authorization service.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var actions []string
	resp.Diagnostics.Append(plan.Actions.ElementsAs(ctx, &actions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateRole(ctx, plan.Name.ValueString(), actions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role",
			fmt.Sprintf("Could not create role: %s", err.Error()),
		)
		return
	}

	plan.ID = plan.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the role from the authorization service and refreshes state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.GetRole(ctx, state.Name.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading role",
			fmt.Sprintf("Could not get role: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(role.Name)
	state.Name = types.StringValue(role.Name)

	actionVals := make([]attr.Value, len(role.Actions))
	for i, a := range role.Actions {
		actionVals[i] = types.StringValue(a)
	}
	actionsList, diags := types.ListValue(types.StringType, actionVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Actions = actionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the role's actions in the authorization service.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var actions []string
	resp.Diagnostics.Append(plan.Actions.ElementsAs(ctx, &actions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.UpdateRole(ctx, plan.Name.ValueString(), actions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating role",
			fmt.Sprintf("Could not update role: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(role.Name)
	plan.Name = types.StringValue(role.Name)

	actionVals := make([]attr.Value, len(role.Actions))
	for i, a := range role.Actions {
		actionVals[i] = types.StringValue(a)
	}
	actionsList, diags := types.ListValue(types.StringType, actionVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Actions = actionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the role from the authorization service.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRole(ctx, state.Name.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			slog.WarnContext(ctx, "role not found during delete, removing from state", "name", state.Name.ValueString())
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting role",
			fmt.Sprintf("Could not delete role: %s", err.Error()),
		)
	}
}

// ImportState imports an existing role by name.
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
