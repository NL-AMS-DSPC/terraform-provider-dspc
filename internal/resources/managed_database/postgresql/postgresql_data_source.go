// Package postgresql implements the Terraform resource and data source for managing PostgreSQL instances in the ASC platform.
package postgresql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines the interface for the methods needed to retrieve PostgreSQL instance information.
type DataClient interface {
	GetPostgreSQLInstance(ctx context.Context, instanceName string) (*client.PostgreSQLInstance, error)
	ListPostgreSQLInstances(ctx context.Context) (*client.ListPostgreSQLInstancesResponse, error)
}

// DataSource implements the Terraform data source for retrieving PostgreSQL instance information from the ASC platform.
type DataSource struct {
	client DataClient
}

// DataSourceModel represents the schema for the PostgreSQL data source in Terraform.
type DataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	Skuize  types.String `tfsdk:"sku_size"`
	Version types.String `tfsdk:"version"`
	VPCID   types.String `tfsdk:"vpc_id"`
	Tags    []TagModel   `tfsdk:"tags"`
}

// NewDataSource creates a new instance of the PostgreSQL data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata returns the data source type name.
func (s *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql"
}

// Schema defines the schema for the PostgreSQL data source, including the required and computed attributes.
func (s *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a PostgreSQL instance in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name of the database instance to retrieve.",
			},
			"sku_size": schema.StringAttribute{
				Computed:    true,
				Description: "Sku size per instance node, e.g. gp-2, gp-4, etc",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the database engine. One of: POSTGRES_15, POSTGRES_16, POSTGRES_17, POSTGRES_18.",
			},
			"vpc_id": schema.StringAttribute{
				Computed:    true,
				Description: "GUID of the VPC network where this database should be added to.",
			},
			"tags": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Tags applied to the database instance.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "Tag key.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "Tag value.",
						},
					},
				},
			},
		},
	}
}

// Configure sets the data source's client based on the provider configuration.
func (s *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Datasource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.ManagedDB == nil {
		resp.Diagnostics.AddError(
			"Unexpected datasource configuration error",
			"Expected a configured ASC client with a ManagedDB client, but the ManagedDB client was nil. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	s.client = dataClient.ManagedDB
}

// Read retrieves the PostgreSQL instance information based on the provided name and updates the Terraform state with the retrieved data.
func (s *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := s.client.GetPostgreSQLInstance(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PostgreSQL instance",
			fmt.Sprintf("An error was encountered when reading the PostgreSQL instance: %s", err.Error()),
		)
		return
	}

	state := toResourceModel(instance)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
