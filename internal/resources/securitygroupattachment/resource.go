// Package securitygroupattachment provides the Terraform resource and data source for
// managing Security Group attachments (attaching/detaching Security Groups to targets such as VMs and Pods).
package securitygroupattachment

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
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing security group attachment resources.
type ResourceClient interface {
	AttachSecurityGroup(ctx context.Context, sgName, targetType, targetName string) (*client.SecurityGroupAttachment, error)
	GetSecurityGroupAttachment(ctx context.Context, sgName, attachmentName string) (*client.SecurityGroupAttachment, error)
	DetachSecurityGroup(ctx context.Context, sgName, attachmentName string) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	TargetType        types.String `tfsdk:"target_type"`
	TargetName        types.String `tfsdk:"target_name"`
	AttachmentName    types.String `tfsdk:"attachment_name"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_attachment"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Security Group attachment to a target resource (e.g. VirtualMachine) in the DSPC platform. " +
			"This attaches a Security Group's network rules to a specific target.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the security group attachment.",
				Computed:    true,
			},
			"security_group_name": schema.StringAttribute{
				Description: "The name of the Security Group to attach.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Description: "The type of target resource to attach to. Valid values: VirtualMachine, Pod.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_name": schema.StringAttribute{
				Description: "The name of the target resource to attach the Security Group to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attachment_name": schema.StringAttribute{
				Description: "The name of the attachment resource created in Kubernetes.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
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

// Create creates a new security group attachment.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sga, err := r.client.AttachSecurityGroup(
		ctx,
		plan.SecurityGroupName.ValueString(),
		plan.TargetType.ValueString(),
		plan.TargetName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching Security Group",
			fmt.Sprintf("Could not attach security group: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(createStateID(plan.SecurityGroupName.ValueString(), sga.Name))
	plan.AttachmentName = types.StringValue(sga.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the security group attachment from the API and stores it in the state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sga, err := r.client.GetSecurityGroupAttachment(
		ctx,
		state.SecurityGroupName.ValueString(),
		state.AttachmentName.ValueString(),
	)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting Security Group Attachment",
			fmt.Sprintf("Could not get security group attachment: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(createStateID(state.SecurityGroupName.ValueString(), sga.Name))
	state.AttachmentName = types.StringValue(sga.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for security group attachments.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Security Group attachment updates are not supported. Changes require recreation.",
	)
}

// Delete detaches the security group.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DetachSecurityGroup(
		ctx,
		state.SecurityGroupName.ValueString(),
		state.AttachmentName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error detaching Security Group",
			fmt.Sprintf("Could not detach security group: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the security group attachment.
// The import ID should be in the format:
// "security-group-name:attachment-name:target-type:target-name"
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 4)
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format 'security-group-name:attachment-name:target-type:target-name', got: %s", req.ID),
		)
		return
	}

	sgName := parts[0]
	attachmentName := parts[1]
	targetType := parts[2]
	targetName := parts[3]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), createStateID(sgName, attachmentName))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("security_group_name"), sgName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attachment_name"), attachmentName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_type"), targetType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target_name"), targetName)...)
}

// createStateID creates a unique identifier for the security group attachment resource.
func createStateID(sgName, attachmentName string) string {
	return sgName + ":" + attachmentName
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not found") ||
		(len(err.Error()) > 14 && err.Error()[:14] == "API error 404:"))
}
