// Package virtualmachine implements the virtual machine data source.
package virtualmachine

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/sku"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataSourceClient defines an interface for interacting with virtual machine data operations.
// ListVMs retrieves a list of virtual machines from the data source.
type DataSourceClient interface {
	ListVMs(ctx context.Context) ([]client.VM, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataSourceClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	VirtualMachines []DataModel `tfsdk:"virtual_machines"`
}

// DataModel represents a single VM in the data source
type DataModel struct {
	URN             types.String   `tfsdk:"urn"`
	Name            types.String   `tfsdk:"name"`
	SKU             sku.Model      `tfsdk:"sku"`
	Status          types.String   `tfsdk:"status"`
	LastError       types.String   `tfsdk:"last_error"`
	Tags            types.Map      `tfsdk:"tags"`
	AttachedVolumes []types.String `tfsdk:"attached_volumes"`
	OS              OSModel        `tfsdk:"os"`
}

// OSModel represents details about the VM OS
type OSModel struct {
	ID           types.String `tfsdk:"id"`
	Family       types.String `tfsdk:"family"`
	Distribution types.String `tfsdk:"distribution"`
	Release      types.String `tfsdk:"release"`
	DisplayName  types.String `tfsdk:"display_name"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	osAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "ID of the OS.",
			Computed:    true,
		},
		"family": schema.StringAttribute{
			Description: "The family of the OS.",
			Computed:    true,
		},
		"distribution": schema.StringAttribute{
			Description: "The distribution of the OS.",
			Computed:    true,
		},
		"release": schema.StringAttribute{
			Description: "The release of the OS.",
			Computed:    true,
		},
		"display_name": schema.StringAttribute{
			Description: "The display name of the OS.",
			Computed:    true,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all virtual machines in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"virtual_machines": schema.ListNestedAttribute{
				Description: "List of virtual machines.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"urn": schema.StringAttribute{
							Description: "The uniform resource name for the virtual machine.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the virtual machine.",
							Computed:    true,
						},
						"sku": schema.SingleNestedAttribute{
							Description: "The SKU of the virtual machine.",
							Computed:    true,
							Attributes:  sku.DataSourceAttributes(),
						},
						"status": schema.StringAttribute{
							Description: "The current status of the virtual machine.",
							Computed:    true,
						},
						"last_error": schema.StringAttribute{
							Description: "The last error encountered during CRUD of the virtual machine.",
							Computed:    true,
						},
						"tags": schema.MapAttribute{
							Description: "User defined tags attached to the virtual machine.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"attached_volumes": schema.ListAttribute{
							Description: "The list of volume names attached to the virtual machine.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"os": schema.SingleNestedAttribute{
							Description: "Details about the virtual machine's OS.",
							Computed:    true,
							Attributes:  osAttrs,
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	if dataClient.VirtualMachines == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected virtual machines service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.VirtualMachines
}

// Read reads the data from the API and stores it in the state.
func (d *DataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel

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
	state.VirtualMachines = make([]DataModel, len(vms))
	for i, vm := range vms {
		attachedVolumes := make([]types.String, len(vm.AttachedVolumes))
		for j, v := range vm.AttachedVolumes {
			attachedVolumes[j] = types.StringValue(v)
		}

		state.VirtualMachines[i] = DataModel{
			URN:             types.StringValue(vm.URN),
			Name:            types.StringValue(vm.Name),
			Status:          types.StringValue(vm.Status),
			LastError:       types.StringValue(vm.LastError),
			Tags:            tags.ToTerraform(ctx, vm.Tags, &resp.Diagnostics),
			AttachedVolumes: attachedVolumes,
			SKU:             sku.ToTerraform(vm.SKU),
			OS:              toOSModel(vm.OS),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// toOSModel converts a client.OSDetails into a terraform OSModel.
func toOSModel(os client.OSDetails) OSModel {
	return OSModel{
		ID:           types.StringValue(os.ID),
		Family:       types.StringValue(os.Family),
		Distribution: types.StringValue(os.Distribution),
		Release:      types.StringValue(os.Release),
		DisplayName:  types.StringValue(os.DisplayName),
	}
}
