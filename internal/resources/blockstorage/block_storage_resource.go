package blockstorage

import (
	"context"
	"fmt"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &BlockResource{}
	_ resource.ResourceWithConfigure   = &BlockResource{}
	_ resource.ResourceWithImportState = &BlockResource{}
)

type blockStorageClient interface {
	UpdateBlock(ctx context.Context, name string, req client.UpdateBlockRequest) (*client.UpdateBlockResponse, error)
	CreateBlock(ctx context.Context, req client.CreateBlockRequest) (*client.CreateBlockResponse, error)
	GetBlock(ctx context.Context, name string) (*client.Block, error)
	DeleteBlock(ctx context.Context, name string) error
}

type BlockResource struct {
	client blockStorageClient
}
type BlockResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.String `tfsdk:"size"`
}

// NewBlockResource creates a new BlockResource
func NewBlockResource() resource.Resource { return &BlockResource{} }

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *BlockResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(blockStorageClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected blockStorageClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = dataClient
}

func (r *BlockResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block"
}

func (r *BlockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Block in the DSPC platform",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the block. Must be unique within the platform",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.StringAttribute{
				Description: "Size of the block storage (e.g. 10Gi)",
				Required:    true,
			},
		},
	}
}

func (r *BlockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlockResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the Block via the API
	block, err := r.client.CreateBlock(ctx, client.CreateBlockRequest{
		Name: plan.Name.ValueString(),
		Size: plan.Size.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Block",
			fmt.Sprintf("Could not create Block: %s", err.Error()),
		)
		return
	}

	// Set the computed values
	plan.ID = types.StringValue(block.Created) // Using name as ID since API doesn't return separate ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state
func (r *BlockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlockResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the Block from the API
	block, err := r.client.GetBlock(ctx, state.Name.ValueString())
	if err != nil {
		// If Block not found, remove from state
		// TODO: could be multiple reasons for failure? why always remove?
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(block.Name)
	state.Name = types.StringValue(block.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BlockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BlockResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the Block via the API
	updateResp, err := r.client.UpdateBlock(ctx, plan.Name.ValueString(), client.UpdateBlockRequest{
		Size: plan.Size.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Block",
			fmt.Sprintf("Could not update Block: %s", err.Error()),
		)
		return
	}

	// Set the computed values
	plan.ID = types.StringValue(updateResp.Name) // Using name as ID since API doesn't return separate ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BlockResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BlockResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the Block via the API
	err := r.client.DeleteBlock(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Block",
			fmt.Sprintf("Could not delete Block: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the block from the DSPC platform.
func (r *BlockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
