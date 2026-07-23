// Package group provides Terraform resources and data sources for managing groups.
package group

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &RoleResource{}
	_ resource.ResourceWithConfigure   = &RoleResource{}
	_ resource.ResourceWithImportState = &RoleResource{}
)

// RoleResourceClient defines the interface for managing group role assignments.
type RoleResourceClient interface {
	AddRoleToGroup(ctx context.Context, groupName, roleName string) error
	RemoveRoleFromGroup(ctx context.Context, groupName, roleName string) error
	GetRolesForGroup(ctx context.Context, groupName string) ([]string, error)
}

// RoleResource defines the resource implementation.
type RoleResource struct {
	client RoleResourceClient
}

// RoleResourceModel describes the resource data model.
type RoleResourceModel struct {
	ID        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	RoleName  types.String `tfsdk:"role_name"`
}

// NewRoleResource creates a new RoleResource.
func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_role"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the assignment of a role to a authorization group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this assignment.",
				Computed:    true,
			},
			"group_name": schema.StringAttribute{
				Description: "The name of the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_name": schema.StringAttribute{
				Description: "The name of the role to assign to the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure stores the provider-configured client on the resource.
func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create assigns a role to a group in the authorization service.
func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddRoleToGroup(ctx, plan.GroupName.ValueString(), plan.RoleName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group role assignment",
			fmt.Sprintf("Could not add role %q to group %q: %s",
				plan.RoleName.ValueString(), plan.GroupName.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(roleCompositeID(plan.GroupName.ValueString(), plan.RoleName.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read checks whether the role is still assigned to the group and removes the resource from state if not.
func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, err := r.client.GetRolesForGroup(ctx, state.GroupName.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading group role assignment",
			fmt.Sprintf("Could not get roles for group %q: %s", state.GroupName.ValueString(), err.Error()),
		)
		return
	}

	for _, role := range roles {
		if role == state.RoleName.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	// Role is no longer assigned — remove from state.
	resp.State.RemoveResource(ctx)
}

// Update is not supported — all fields require replacement.
func (r *RoleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Group role assignment updates are not supported. Changes require destroying and recreating the assignment.",
	)
}

// Delete removes a role from a group in the authorization service.
func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveRoleFromGroup(ctx, state.GroupName.ValueString(), state.RoleName.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting group role assignment",
			fmt.Sprintf("Could not remove role %q from group %q: %s",
				state.RoleName.ValueString(), state.GroupName.ValueString(), err.Error()),
		)
	}
}

// ImportState imports an existing group role assignment using the "group_name:role_name" format.
func (r *RoleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in format \"group_name:role_name\", got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_name"), parts[1])...)
}

func roleCompositeID(groupName, roleName string) string {
	return fmt.Sprintf("%s:%s", groupName, roleName)
}
