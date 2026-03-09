// Package role provides Terraform resources and data sources for managing roles.
package role

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
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
	Name    types.String `tfsdk:"name"`
	Actions types.List   `tfsdk:"actions"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a role and its associated permissions from the authorization service.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the role to look up.",
				Required:    true,
			},
			"actions": schema.ListAttribute{
				Description: "The list of permission actions assigned to this role.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure stores the provider-configured client on the data source.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dspcClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dspcClient.Authorization == nil {
		resp.Diagnostics.AddError(
			"Unexpected data source configuration error",
			"Expected authorization service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dspcClient.Authorization
}

// Read reads the role from the authorization service.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRole(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			fmt.Sprintf("Could not get role %q: %s", config.Name.ValueString(), err.Error()),
		)
		return
	}

	actionVals := make([]attr.Value, len(role.Actions))
	for i, a := range role.Actions {
		actionVals[i] = types.StringValue(a)
	}
	actionsList, diags := types.ListValue(types.StringType, actionVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Name = types.StringValue(role.Name)
	config.Actions = actionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
