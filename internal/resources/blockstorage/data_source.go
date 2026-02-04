// Package blockstorage implements the block storage data source.
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
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

type blockDataClient interface {
	GetBlock(ctx context.Context, name string) (*client.Block, error)
}

// DataSourceModel describes the resource data model
type DataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Size types.String `tfsdk:"size"`
}

// DataSource defines the resource implementation.
type DataSource struct {
	client blockDataClient
}

// Metadata updates the provided metadata with the resource type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a specific block in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the block.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the block.",
				Required:    true,
			},
			"size": schema.StringAttribute{
				Description: "The size of the block.",
				Computed:    true,
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *DataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.BlockStorage == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected blockstorage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.BlockStorage
}

// Read reads the data from the API and stores it in the state.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get requested Block from the API
	block, err := d.client.GetBlock(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Block",
			fmt.Sprintf("Could not get block: %s", err.Error()),
		)
		return
	}

	dataSourceBlock := DataSourceModel{
		ID:   types.StringValue(block.Name),
		Name: types.StringValue(block.Name),
		Size: types.StringValue(block.Size),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataSourceBlock)...)
}

// NewDataSource creates a new DataSource
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}
