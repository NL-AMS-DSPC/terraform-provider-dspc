package resources

import (
	"context"

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
	CreateAttachment(ctx context.Context, pvcName, vmName string) (*client.CreateBlockAttachmentResponse, error)
}

type BlockStorageAttachmentResource struct {
	client BlockStorageAttachmentClient
}

type BlockStorageAttachmentResourceModel struct {
	Name         types.String `tfsdk:"name" example:"my-pvc"`
	Size         types.String `tfsdk:"size" example:"10Gi"`
	StorageClass types.String `tfsdk:"storageClass" example:"standard"`
	AccessMode   types.String `tfsdk:"accessMode" example:"ReadWriteOnce" enum:"ReadWriteOnce,ReadWriteMany,ReadOnlyMany"`
	VolumeMode   types.String `tfsdk:"volumeMode" example:"Filesystem" enum:"Filesystem,Block"`
	Status       types.String `tfsdk:"status" example:"Bound" enum:"Pending,Bound,Lost"`
	Namespace    types.String `tfsdk:"namespace,omitempty" example:"default"`
	AttachedToVM types.String `tfsdk:"attachedToVM,omitempty" example:"my-vm"`
	Labels       types.Map    `tfsdk:"labels,omitempty"`
	Annotations  types.Map    `tfsdk:"annotations,omitempty"`
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
	//TODO implement me
	panic("implement me")
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
