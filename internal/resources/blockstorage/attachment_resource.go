package blockstorage

import (
	"context"
	"errors"
	"fmt"

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
	_ resource.Resource                = &AttachmentResource{}
	_ resource.ResourceWithConfigure   = &AttachmentResource{}
	_ resource.ResourceWithImportState = &AttachmentResource{}
)

// AttachmentClient defines the interface for managing block storage attachment resources.
// It provides methods to create, delete, and retrieve block storage attachments.
type AttachmentClient interface {
	CreateAttachment(ctx context.Context, blockName, vmName string) (*client.BlockStorageAttachment, error)
	GetAttachment(ctx context.Context, blockName, vmName string) (*client.BlockStorageAttachment, error)
	DeleteAttachment(ctx context.Context, blockName, vmName string) error
}

// AttachmentResource defines the resource implementation.
type AttachmentResource struct {
	client AttachmentClient
}

// AttachmentResourceModel describes the resource data model.
type AttachmentResourceModel struct {
	ID               types.String `tfsdk:"id" example:"my-block-storage-my-vm"`
	VMName           types.String `tfsdk:"vm_name" example:"my-vm"`
	BlockStorageName types.String `tfsdk:"block_storage_name" example:"my-block-storage"`
}

// NewAttachmentResource creates a new AttachmentResource.
func NewAttachmentResource() resource.Resource {
	return &AttachmentResource{}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (b *AttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected AttachmentClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	if c.BlockStorage == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected blockstorage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	b.client = c.BlockStorage
}

// Metadata updates the provided metadata with the resource type name.
func (b *AttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage_attachment"
}

// Schema updates the resource schema with the attributes for the resource.
func (b *AttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a block storage attachment",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the block storage attachment",
				Computed:    true,
			},
			"vm_name": schema.StringAttribute{
				Description: "The name of the virtual machine. Must be unique within the platform.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"block_storage_name": schema.StringAttribute{
				Description: "The name of the block storage. Must be unique within the platform.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create creates a new block storage attachment in the DSPC platform.
func (b *AttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AttachmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	attachment, err := b.client.CreateAttachment(ctx, plan.BlockStorageName.ValueString(), plan.VMName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating block storage attachment",
			fmt.Sprintf("Error creating block storage attachment: %s", err.Error()))
		return
	}

	// Set computed values
	plan.ID = types.StringValue(createStateID(attachment.BlockName, attachment.VMName)) // using the combination of names as ID since the API doesn't provide an ID at this point in time
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (b *AttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the attachment from the API
	attachment, err := b.client.GetAttachment(ctx, state.BlockStorageName.ValueString(), state.VMName.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			// If attachment not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting Block attachment",
			fmt.Sprintf("Could not get Block attachment: %s", err.Error()),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(createStateID(attachment.BlockName, attachment.VMName)) // using the combination of names as ID since the API doesn't provide an ID at this point in time
	state.BlockStorageName = types.StringValue(attachment.BlockName)
	state.VMName = types.StringValue(attachment.VMName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the block storage attachment in the DSPC platform.
func (b *AttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// We only support the creation and deletion of attachments, so an update isn't available.
	resp.Diagnostics.AddError(
		"Update not supported",
		"Block storage attachments updates are not supported by the DSPC API. Changes require attachment recreation. "+
			"Consider using lifecycle { ignore_changes = [name] } if you need to prevent replacement.",
	)
}

// Delete deletes the block storage attachment in the DSPC platform.
func (b *AttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AttachmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the VM via the API
	err := b.client.DeleteAttachment(ctx, state.BlockStorageName.ValueString(), state.VMName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting block storage attachment",
			fmt.Sprintf("Could not delete block storage attachment: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the block storage attachment from the DSPC platform.
// The import ID should be in the format: "block-storage-name:vm-name"
func (b *AttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Split the import ID into block storage name and VM name
	parts := splitImportID(req.ID)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format 'block-storage-name:vm-name', got: %s", req.ID),
		)
		return
	}

	blockStorageName := parts[0]
	vmName := parts[1]

	// Set the individual attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), createStateID(blockStorageName, vmName))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("block_storage_name"), blockStorageName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vm_name"), vmName)...)
}

// createStateID creates a unique identifier for the block storage attachment resource.
func createStateID(blockName, vmName string) string {
	return blockName + ":" + vmName
}

// splitImportID splits an import ID string by colon separator.
func splitImportID(id string) []string {
	parts := make([]string, 0, 2)
	for i := 0; i < len(id); i++ {
		if id[i] == ':' {
			if len(parts) == 0 {
				parts = append(parts, id[:i])
				if i+1 < len(id) {
					parts = append(parts, id[i+1:])
				}
			}
			break
		}
	}
	if len(parts) == 0 {
		parts = append(parts, id)
	}
	return parts
}
