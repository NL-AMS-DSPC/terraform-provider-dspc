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
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &MemberResource{}
	_ resource.ResourceWithConfigure   = &MemberResource{}
	_ resource.ResourceWithImportState = &MemberResource{}
)

// MemberResourceClient defines the interface for managing group membership.
type MemberResourceClient interface {
	AddUserToGroup(ctx context.Context, groupName, userID string) error
	RemoveUserFromGroup(ctx context.Context, groupName, userID string) error
}

// MemberResource defines the resource implementation.
type MemberResource struct {
	client MemberResourceClient
}

// MemberResourceModel describes the resource data model.
type MemberResourceModel struct {
	ID        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	UserID    types.String `tfsdk:"user_id"`
}

// NewMemberResource creates a new MemberResource.
func NewMemberResource() resource.Resource {
	return &MemberResource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *MemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_member"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *MemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the membership of a user in a authorization group. " +
			"Note: the authorization API does not expose a read endpoint for group membership, " +
			"so drift cannot be detected — the resource relies on Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this membership, in the format \"group_name:user_id\".",
				Computed:    true,
			},
			"group_name": schema.StringAttribute{
				Description: "The name of the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to add to the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure stores the provider-configured client on the resource.
func (r *MemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dspcClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dspcClient.Authorization == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected authorization service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dspcClient.Authorization
}

// Create adds a user to a group in the authorization service.
func (r *MemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MemberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddUserToGroup(ctx, plan.GroupName.ValueString(), plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group membership",
			fmt.Sprintf("Could not add user %q to group %q: %s",
				plan.UserID.ValueString(), plan.GroupName.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(compositeID(plan.GroupName.ValueString(), plan.UserID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read is a no-op because the authorization API does not expose a read endpoint
// for group membership. State is assumed to be accurate.
func (r *MemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Membership cannot be verified via the HTTP API; preserve existing state.
}

// Update is not supported — all fields require replacement.
func (r *MemberResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Group membership updates are not supported. Changes require destroying and recreating the membership.",
	)
}

// Delete removes a user from a group in the authorization service.
func (r *MemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MemberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.RemoveUserFromGroup(ctx, state.GroupName.ValueString(), state.UserID.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting group membership",
			fmt.Sprintf("Could not remove user %q from group %q: %s",
				state.UserID.ValueString(), state.GroupName.ValueString(), err.Error()),
		)
	}
}

// ImportState imports an existing group membership using the "group_name:user_id" format.
func (r *MemberResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in format \"group_name:user_id\", got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), parts[1])...)
}

func compositeID(groupName, userID string) string {
	return fmt.Sprintf("%s:%s", groupName, userID)
}
