package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &VirtualMachineResource{}
	_ resource.ResourceWithConfigure   = &VirtualMachineResource{}
	_ resource.ResourceWithImportState = &VirtualMachineResource{}
)

// VirtualMachineResource defines the resource implementation.
type VirtualMachineResource struct{}

// VirtualMachineResourceModel describes the resource data model.
type VirtualMachineResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewVirtualMachineResource() resource.Resource {
	return &VirtualMachineResource{}
}

func (r *VirtualMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *VirtualMachineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the virtual machine.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the virtual machine. Must be unique within the platform.",
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The timestamp when the virtual machine was created.",
				Computed:    true,
			},
		},
	}
}

func (r *VirtualMachineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Empty implementation for docs generation
}

func (r *VirtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Empty implementation for docs generation
}

func (r *VirtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Empty implementation for docs generation
}

func (r *VirtualMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Empty implementation for docs generation
}

func (r *VirtualMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Empty implementation for docs generation
}

func (r *VirtualMachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
