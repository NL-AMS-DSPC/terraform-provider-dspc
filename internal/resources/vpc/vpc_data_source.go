package vpc

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
	_ datasource.DataSource              = &VPCDataSource{}
	_ datasource.DataSourceWithConfigure = &VPCDataSource{}
)

// VPCDataClient defines an interface for interacting with VPC data operations.
type VPCDataClient interface {
	ListVPCs(ctx context.Context) ([]*client.VPC, error)
}

// VPCDataSource defines the data source implementation.
type VPCDataSource struct {
	client VPCDataClient
}

// VPCDataSourceModel describes the data source data model.
type VPCDataSourceModel struct {
	VPCs []VPCModel `tfsdk:"vpcs"`
}

// VPCModel represents a single VPC in the data source.
type VPCModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	CIDR   types.String `tfsdk:"cidr"`
	Status types.String `tfsdk:"status"`
}

// NewVPCDataSource creates a new VPCDataSource.
func NewVPCDataSource() datasource.DataSource {
	return &VPCDataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *VPCDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *VPCDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all VPCs in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"vpcs": schema.ListNestedAttribute{
				Description: "List of VPCs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the VPC.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the VPC.",
							Computed:    true,
						},
						"cidr": schema.StringAttribute{
							Description: "The CIDR range of the VPC.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the VPC.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *VPCDataSource) Configure(
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
func (d *VPCDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VPCDataSourceModel

	vpcs, err := d.client.ListVPCs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing VPCs",
			fmt.Sprintf("Could not list VPCs: %s", err.Error()),
		)
		return
	}

	state.VPCs = make([]VPCModel, len(vpcs))
	for i, v := range vpcs {
		state.VPCs[i] = VPCModel{
			ID:     types.StringValue(v.Name),
			Name:   types.StringValue(v.Name),
			CIDR:   types.StringValue(v.CIDR),
			Status: types.StringValue(v.Status),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
