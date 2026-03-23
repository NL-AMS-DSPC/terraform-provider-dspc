package mssql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

type DataClient interface {
	GetMSSQLInstance(ctx context.Context, instanceName string) (*client.MSSQLInstance, error)
	ListMSSQLInstances(ctx context.Context) (*client.ListMSSQLInstancesResponse, error)
}

type DataSource struct {
	client DataClient
}

type DataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	Size    types.String `tfsdk:"size"`
	Version types.String `tfsdk:"version"`
	VPC     types.String `tfsdk:"vpc"`
	Tags    []TagModel   `tfsdk:"tags"`
}

func NewMSSQLSource() *DataSource {
	return &DataSource{}
}

func (s *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mssql"
}

func (s *DataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a MSSQL instance in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name of the database instance to retrieve.",
			},
			"size": schema.StringAttribute{
				Computed:    true,
				Description: "Size of the database storage, e.g. 500Mi, 1Gi.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the database engine. One of: MSSQL_2025_17, MSSQL_2022_16, MSSQL_2019_15, MSSQL_2017_14.",
			},
			"vpc": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the VPC network where this database is deployed.",
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

	if dataClient.Network == nil {
		resp.Diagnostics.AddError(
			"Unexpected datasource configuration error",
			"Expected a configured DSPC client with a Network client, but the Network client was nil. "+
				"Please report this issue to the provider developers.",
		)
		return
	}

	s.client = dataClient.Network
}

func (s *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := s.client.GetMSSQLInstance(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MSSQL instance",
			fmt.Sprintf("An error was encountered when reading the MSSQL instance: %s", err.Error()),
		)
		return
	}

	state := toResourceModel(instance)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
