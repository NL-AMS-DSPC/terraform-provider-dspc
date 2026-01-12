package blockstorage

import (
	"context"
	"fmt"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &BlockStorageDataSource{}
	_ datasource.DataSourceWithConfigure = &BlockStorageDataSource{}
)

type blockDataClient interface {
	ListBlocks(ctx context.Context) ([]*client.Block, error)
}

type BlockStorageDataSource struct {
	client blockDataClient
}

func (d *BlockStorageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blocks"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *BlockStorageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all blocks in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"blocks": schema.ListNestedAttribute{
				Description: "List of blocks.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the block.",
							Computed:    true,
						},
						"size": schema.StringAttribute{
							Description: "The size of the block",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *BlockStorageDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(blockDataClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected blockDataClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = dataClient
}

// Read reads the data from the API and stores it in the state.
func (d *BlockStorageDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state BlockStorageDataSourceModel

	// Get all Blocks from the API
	blocks, err := d.client.ListBlocks(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing Blocks",
			fmt.Sprintf("Could not list Blocks: %s", err.Error()),
		)
		return
	}

	// Convert API Blocks to Terraform model
	state.Blocks = make([]BlockModel, len(blocks))
	for i, block := range blocks {
		state.Blocks[i] = BlockModel{
			ID:   types.StringValue(block.Name),
			Size: types.StringValue(block.Size),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type BlockModel struct {
	ID   types.String `tfsdk:"id"`
	Size types.String `tfsdk:"size"`
}

type BlockStorageDataSourceModel struct {
	Blocks []BlockModel `tfsdk:"blocks"`
}

func NewBlockDataSource() datasource.DataSource {
	return &BlockStorageDataSource{}
}
