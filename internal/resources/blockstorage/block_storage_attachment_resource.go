package resources

import (
	"context"
	"fmt"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &BlockStorageAttachmentResource{}
	_ resource.ResourceWithConfigure   = &BlockStorageAttachmentResource{}
	_ resource.ResourceWithImportState = &BlockStorageAttachmentResource{}
)

type BlockStorageAttachmentClient interface {
	CreateAttachment(ctx context.Context, blockName, vmName string) (*client.BlockStorageAttachment, error)
}

type BlockStorageAttachmentResource struct {
	client BlockStorageAttachmentClient
}

type BlockStorageAttachmentResourceModel struct {
	ID               types.String `tfsdk:"id" example:"my-block-storage-my-vm"`
	VMName           types.String `tfsdk:"vm_name" example:"my-vm"`
	BlockStorageName types.String `tfsdk:"block_storage_name" example:"my-block-storage"`
}

func NewBlockStorageAttachmentResource() resource.Resource {
	return &BlockStorageAttachmentResource{}
}

func (b *BlockStorageAttachmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockStorageAttachmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage_attachment"
}

func (b *BlockStorageAttachmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (b *BlockStorageAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockStorageAttachmentResourceModel
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
	plan.ID = types.StringValue(fmt.Sprintf("%s-%s", attachment.BlockName, attachment.VMName)) // using the combination of names as ID since the API doesn't provide an ID at this point in time
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (b *BlockStorageAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockStorageAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockStorageAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (b *BlockStorageAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	//TODO implement me
	panic("implement me")
}
