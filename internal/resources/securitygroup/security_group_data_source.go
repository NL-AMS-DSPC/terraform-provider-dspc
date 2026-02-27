// Package securitygroup provides Terraform resources and data sources for managing Security Groups.
package securitygroup

import (
	"context"
	"fmt"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines an interface for interacting with Security Group data operations.
type DataClient interface {
	ListSecurityGroups(ctx context.Context) ([]*client.SecurityGroup, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	SecurityGroups []SecurityGroupModel `tfsdk:"security_groups"`
}

// SecurityGroupModel represents a single Security Group in the data source.
type SecurityGroupModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all Security Groups in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"security_groups": schema.ListNestedAttribute{
				Description: "List of Security Groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the Security Group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the Security Group.",
							Computed:    true,
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

	sgs, err := d.client.ListSecurityGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing Security Groups",
			fmt.Sprintf("Could not list Security Groups: %s", err.Error()),
		)
		return
	}

	state.SecurityGroups = make([]SecurityGroupModel, len(sgs))
	for i, sg := range sgs {
		state.SecurityGroups[i] = SecurityGroupModel{
			ID:   types.StringValue(sg.Name),
			Name: types.StringValue(sg.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
