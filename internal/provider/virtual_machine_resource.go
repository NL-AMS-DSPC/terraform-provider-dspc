package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &VMResource{}
	_ resource.ResourceWithConfigure   = &VMResource{}
	_ resource.ResourceWithImportState = &VMResource{}
)

// VMResource defines the resource implementation.
type VMResource struct {
	client *Client
}

// VMResourceModel describes the resource data model.
type VMResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewVMResource creates a new VMResource.
func NewVMResource() resource.Resource {
	return &VMResource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *VMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *VMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *VMResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new virtual machine in the DSPC platform.
func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the VM via the API
	vm, err := r.client.CreateVM(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VM",
			fmt.Sprintf("Could not create VM: %s", err.Error()),
		)
		return
	}

	// Set the computed values
	plan.ID = types.StringValue(vm.Name) // Using name as ID since API doesn't return separate ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the VM from the API
	vm, err := r.client.GetVM(ctx, state.Name.ValueString())
	if err != nil {
		// If VM not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(vm.Name)
	state.Name = types.StringValue(vm.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the virtual machine in the DSPC platform.
func (r *VMResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since the API only supports VM name and doesn't have update operations,
	// we treat any changes as requiring recreation (ForceNew)
	resp.Diagnostics.AddError(
		"Update not supported",
		"VM updates are not supported by the DSPC API. Changes require VM recreation. "+
			"Consider using lifecycle { ignore_changes = [name] } if you need to prevent replacement.",
	)
}

// Delete deletes the virtual machine in the DSPC platform.
func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the VM via the API
	err := r.client.DeleteVM(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VM",
			fmt.Sprintf("Could not delete VM: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the virtual machine in the DSPC platform.
func (r *VMResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
