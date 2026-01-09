package resources

import (
	"context"
	"fmt"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
// var (
//
//	_ resource.Resource                = &VMResource{}
//	_ resource.ResourceWithConfigure   = &VMResource{}
//	_ resource.ResourceWithImportState = &VMResource{}
//
// )
// TODO: name BlockStorage

type blockStorageClient interface {
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

func (r *BlockResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block"
}

func (r *BlockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Block in the DSPC platform",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{},
			"size": schema.StringAttribute{},
		},
	}
}

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
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(block.Name)
	state.Name = types.StringValue(block.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BlockResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me. support resize update here or through vm?
	panic("implement me")
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

type BlockModel struct {
	ID   types.String `tfsdk:"id"`
	Size types.Int64  `tfsdk:"size"`
}

// name
// size
// storageClass
// accessMode ->
// volumeMode
