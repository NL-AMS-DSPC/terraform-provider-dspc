package blockstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type bsClient interface {
	UpdateBlock(ctx context.Context, req client.UpdateBlockRequest) (*client.UpdateBlockResponse, error)
	CreateBlock(ctx context.Context, req client.CreateBlockRequest) (*client.CreateBlockResponse, error)
	GetBlock(ctx context.Context, name string) (*client.Block, error)
	DeleteBlock(ctx context.Context, name string) error
}

type blockResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.String `tfsdk:"size"`
}

// Resource defines the resource implementation.
type Resource struct {
	client bsClient
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.BlockStorage == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected blockstorage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.BlockStorage
}

// Metadata returns the full name of the resource
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage"
}

// Schema returns the schema for this resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Block in the ASC platform",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the block. Must be unique within the platform",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the block. Must be unique within the platform",
				Required:    true,
			},
			"size": schema.StringAttribute{
				Description: "Size of the block storage (e.g. 10Gi)",
				Required:    true,
			},
		},
	}
}

// Create will create the resource and set the initial Terraform state
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blockResourceModel

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
	plan.ID = types.StringValue(block.Created)
	plan.Name = types.StringValue(block.Created) // Using name as ID since API doesn't return separate ID
	plan.Size = types.StringValue(plan.Size.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blockResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the Block from the API
	block, err := r.client.GetBlock(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			// If Block not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting Block",
			fmt.Sprintf("Could not get Block: %s", err.Error()),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(block.Name)
	state.Name = types.StringValue(block.Name)
	state.Size = types.StringValue(block.Size)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is called to update the state of the resource
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan blockResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update the Block via the API
	updateResp, err := r.client.UpdateBlock(ctx, client.UpdateBlockRequest{
		Name: plan.Name.ValueString(),
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
	plan.ID = types.StringValue(updateResp.Name)
	plan.Name = types.StringValue(updateResp.Name) // Using name as ID since API doesn't return separate ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is called when the provider must delete the resource
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blockResourceModel

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

// ImportState imports the state of the block from the ASC platform.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NewBlockStorageResource creates a new Resource
func NewBlockStorageResource() resource.Resource { return &Resource{} }
