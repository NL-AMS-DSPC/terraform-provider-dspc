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
	_ datasource.DataSource              = &VirtualMachineDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualMachineDataSource{}
)

// VirtualMachineDataSource defines the data source implementation.
type VirtualMachineDataSource struct {
	client *Client
}

// VirtualMachineDataSourceModel describes the data source data model.
type VirtualMachineDataSourceModel struct {
	VirtualMachines []VirtualMachineModel `tfsdk:"virtual_machines"`
}

// VirtualMachineModel represents a single VM in the data source
type VirtualMachineModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewVirtualMachineDataSource() datasource.DataSource {
	return &VirtualMachineDataSource{}
}

func (d *VirtualMachineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

func (d *VirtualMachineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *VirtualMachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VirtualMachineDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VirtualMachineDataSourceModel

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
	state.VirtualMachines = make([]VirtualMachineModel, len(vms))
	for i, vm := range vms {
		state.VirtualMachines[i] = VirtualMachineModel{
			ID:   types.StringValue(vm.Name),
			Name: types.StringValue(vm.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
