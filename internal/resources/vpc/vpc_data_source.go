// Package vpc provides Terraform resources and data sources for managing VPCs.
package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/subnet"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines an interface for interacting with VPC data operations.
type DataClient interface {
	ListVPCs(ctx context.Context) ([]*client.VPC, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	VPCs []ResourceModel `tfsdk:"vpcs"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all VPCs in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"vpcs": schema.ListNestedAttribute{
				Description: "List of VPCs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Identifier for the VPC.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the VPC.",
							Computed:    true,
						},
						"urn": schema.StringAttribute{
							Description: "The uniform resource name of the VPC.",
							Computed:    true,
						},
						"cidr": schema.StringAttribute{
							Description: "The CIDR of the VPC.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The status of the VPC.",
							Computed:    true,
						},
						"last_error": schema.StringAttribute{
							Description: "The last error during CRUD of the VPC.",
							Computed:    true,
						},
						"tags": schema.MapAttribute{
							Description: "Customer-managed key/value tags.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"subnets": schema.ListNestedAttribute{
							Description: "Subnets to create inline as part of the VPC.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: subnet.DatasourceAttributes(),
							},
						},
					},
				},
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

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Network
}

// Read reads the data from the API and stores it in the state.
func (d *DataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel

	vpcs, err := d.client.ListVPCs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing VPCs",
			fmt.Sprintf("Could not list VPCs: %s", err.Error()),
		)
		return
	}

	state.VPCs = make([]ResourceModel, len(vpcs))
	for i, v := range vpcs {
		state.VPCs[i] = ToTerraform(ctx, *v, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
