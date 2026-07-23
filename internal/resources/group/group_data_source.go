// Package group provides Terraform resources and data sources for managing groups.
package group

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataSource defines the data source implementation.
type DataSource struct {
	client ResourceClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a group from the authorization service.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the group to look up.",
				Required:    true,
			},
		},
	}
}

// Configure stores the provider-configured client on the data source.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	ascClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if ascClient.Authorization == nil {
		resp.Diagnostics.AddError(
			"Unexpected data source configuration error",
			"Expected authorization service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = ascClient.Authorization
}

// Read reads the group from the authorization service.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := d.client.GetGroup(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading group",
			fmt.Sprintf("Could not get group %q: %s", config.Name.ValueString(), err.Error()),
		)
		return
	}

	config.Name = types.StringValue(g.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
