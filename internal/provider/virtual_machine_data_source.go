package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &VMDataSource{}
	_ datasource.DataSourceWithConfigure = &VMDataSource{}
)

// VMDataSource defines the data source implementation.
type VMDataSource struct {
	client *Client
}

// VMDataSourceModel describes the data source data model.
type VMDataSourceModel struct {
	VirtualMachines []VMModel `tfsdk:"virtual_machines"`
}

// VMModel represents a single VM in the data source
type VMModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewVMDataSource creates a new VMDataSource.
func NewVMDataSource() datasource.DataSource {
	return &VMDataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *VMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *VMDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all virtual machines in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"virtual_machines": schema.ListNestedAttribute{
				Description: "List of virtual machines.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the virtual machine.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the virtual machine.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *VMDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read reads the data from the API and stores it in the state.
func (d *VMDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VMDataSourceModel

	// Get all VMs from the API
	vms, err := d.client.ListVMs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing VMs",
			fmt.Sprintf("Could not list VMs: %s", err.Error()),
		)
		return
	}

	// Convert API VMs to Terraform model
	state.VirtualMachines = make([]VMModel, len(vms))
	for i, vm := range vms {
		state.VirtualMachines[i] = VMModel{
			ID:   types.StringValue(vm.Name),
			Name: types.StringValue(vm.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
