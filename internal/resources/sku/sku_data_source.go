// Package sku implements the sku data source.
package sku

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines an interface for interacting with SKU data operations.
type DataClient interface {
	ListSKUs(ctx context.Context) ([]client.SKUResponse, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	SKUs []Model `tfsdk:"skus"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_skus"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all available SKUs in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"skus": schema.ListNestedAttribute{
				Description: "List of available SKUs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: DataSourceAttributes(),
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	if dataClient.SKUs == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected skus service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.SKUs
}

// Read reads the data from the API and stores it in the state.
func (d *DataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel

	skus, err := d.client.ListSKUs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing SKUs",
			fmt.Sprintf("Could not list SKUs: %s", err.Error()),
		)
		return
	}

	state.SKUs = make([]Model, len(skus))
	for i, s := range skus {
		state.SKUs[i] = ToTerraform(s)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
