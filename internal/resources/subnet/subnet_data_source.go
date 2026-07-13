// Package subnet provides Terraform resources and data sources for managing subnets.
package subnet

import (
	"context"
	"fmt"

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

// DataClient defines the interface for listing subnets.
type DataClient interface {
	ListSubnetsForVPC(ctx context.Context, vpcName string) ([]*client.Subnet, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	VPCName types.String    `tfsdk:"vpc_name"`
	Subnets []ResourceModel `tfsdk:"subnets"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

// Schema updates the data source schema with the attributes.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of subnets for a VPC from the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"vpc_name": schema.StringAttribute{
				Description: "The name of the VPC to list subnets for.",
				Required:    true,
			},
			"subnets": schema.ListNestedAttribute{
				Description: "The list of subnets in the VPC.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: DatasourceAttributes(),
				},
			},
		},
	}
}

// DatasourceAttributes return the subnet terraform schema attributes
func DatasourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique identifier for the subnet (uuid).",
			Computed:    true,
		},
		"urn": schema.StringAttribute{
			Description: "The uniform resource name for the subnet.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of the subnet. Must be unique within the VPC.",
			Computed:    true,
		},
		"cidr": schema.StringAttribute{
			Description: "The CIDR range for the subnet (e.g. \"10.0.0.0/25\").",
			Computed:    true,
		},
		"type": schema.StringAttribute{
			Description: "The type of the subnet: \"public\" or \"private\".",
			Computed:    true,
		},
		"vpc_name": schema.StringAttribute{
			Description: "The name of the parent VPC.",
			Computed:    true,
		},
		"vpc_id": schema.StringAttribute{
			Description: "The id of the parent VPC.",
			Computed:    true,
		},
		"status": schema.StringAttribute{
			Description: "The current status of the subnet (pending, active, deleting, error).",
			Computed:    true,
		},
		"last_error": schema.StringAttribute{
			Description: "The last error encountered during subnet CRUD operations.",
			Computed:    true,
		},
		"tags": schema.MapAttribute{
			Description: "Customer-managed key/value tags.",
			Computed:    true,
			ElementType: types.StringType,
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected data source configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Network
}

// Read fetches the list of subnets for a VPC and stores them in the state.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vpcName := config.VPCName.ValueString()

	subnets, err := d.client.ListSubnetsForVPC(ctx, vpcName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing subnets",
			fmt.Sprintf("Could not list subnets for VPC %q: %s", vpcName, err.Error()),
		)
		return
	}

	config.Subnets = make([]ResourceModel, len(subnets))
	for i, s := range subnets {
		config.Subnets[i] = SubnetToTerraform(ctx, *s, &resp.Diagnostics)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
