package provider

import (
	"context"

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
type VirtualMachineDataSource struct{}

// VirtualMachineDataSourceModel describes the data source data model.
type VirtualMachineDataSourceModel struct {
	VirtualMachines []VirtualMachineModel `tfsdk:"virtual_machines"`
}

// VirtualMachineModel represents a single VM in the data source
type VirtualMachineModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
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
						"created_at": schema.StringAttribute{
							Description: "The timestamp when the virtual machine was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *VirtualMachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Empty implementation for docs generation
}

func (d *VirtualMachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Empty implementation for docs generation
}
