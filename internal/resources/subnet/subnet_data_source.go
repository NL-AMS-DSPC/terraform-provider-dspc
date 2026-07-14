// Package subnet provides Terraform resources and data sources for managing subnets.
package subnet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
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
	VPCName types.String `tfsdk:"vpc_name"`
	Subnets []Model      `tfsdk:"subnets"`
}

// Model describes a single subnet in the data source.
type Model struct {
	URN       types.String `tfsdk:"urn"`
	Name      types.String `tfsdk:"name"`
	CIDR      types.String `tfsdk:"cidr"`
	Type      types.String `tfsdk:"type"`
	VPCID     types.String `tfsdk:"vpc_id"`
	Status    types.String `tfsdk:"status"`
	LastError types.String `tfsdk:"last_error"`
	Tags      types.Map    `tfsdk:"tags"`
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
					Attributes: map[string]schema.Attribute{
						"urn": schema.StringAttribute{
							Description: "The uniform resource name of the subnet.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the subnet.",
							Computed:    true,
						},
						"cidr": schema.StringAttribute{
							Description: "The CIDR range of the subnet.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the subnet (public or private).",
							Computed:    true,
						},
						"vpc_id": schema.StringAttribute{
							Description: "The ID of the parent VPC.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the subnet.",
							Computed:    true,
						},
						"last_error": schema.StringAttribute{
							Description: "The last error encountered during CRUD of the subnet.",
							Computed:    true,
						},
						"tags": schema.MapAttribute{
							Description: "User defined tags attached to the subnet.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
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

	config.Subnets = make([]Model, len(subnets))
	for i, s := range subnets {
		config.Subnets[i] = Model{
			URN:       types.StringValue(s.URN),
			Name:      types.StringValue(s.Name),
			CIDR:      types.StringValue(s.CIDR),
			Type:      types.StringValue(s.Type),
			VPCID:     types.StringValue(s.VPCID),
			Status:    types.StringValue(s.Status),
			LastError: types.StringValue(s.LastError),
			Tags:      tags.ToTerraform(ctx, s.Tags, &resp.Diagnostics),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
