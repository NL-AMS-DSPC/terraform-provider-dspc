package function

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
	_ datasource.DataSource              = &FunctionDataSource{}
	_ datasource.DataSourceWithConfigure = &FunctionDataSource{}
)

// FunctionDataSourceClient defines the interface for retrieving function data source information.
type FunctionDataSourceClient interface {
	GetFunction(ctx context.Context, name string) (*client.Function, error)
}

// FunctionDataSource defines the data source implementation.
type FunctionDataSource struct {
	client FunctionDataSourceClient
}

// FunctionDataSourceModel describes the data source data model.
type FunctionDataSourceModel struct {
	Name   types.String `tfsdk:"name"`
	ID     types.String `tfsdk:"id"`
	Image  types.String `tfsdk:"image"`
	Status types.String `tfsdk:"status"`
}

// NewFunctionDataSource creates a new FunctionDataSource.
func NewFunctionDataSource() datasource.DataSource {
	return &FunctionDataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *FunctionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *FunctionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific function in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the function to retrieve.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the function.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "The container image for the function.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the function.",
				Computed:    true,
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *FunctionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	if dataClient.Functions == nil {
		resp.Diagnostics.AddError("Unexpected data source configuration error",
			"Expected functions service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Functions
}

// Read refreshes the Terraform state with the latest data.
func (d *FunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FunctionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the function using the provider-level namespace
	function, err := d.client.GetFunction(ctx, config.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting function",
			fmt.Sprintf("Could not get function '%s': %s", config.Name.ValueString(), err.Error()),
		)
		return
	}

	// Set the computed values
	config.ID = types.StringValue(function.Name)
	config.Image = types.StringValue("") // Image not available from API response
	config.Status = types.StringValue(function.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
