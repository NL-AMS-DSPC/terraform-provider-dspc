package sku

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Model represents basic SKU information.
type Model struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Family      types.String `tfsdk:"family"`
	Threads     types.Int64  `tfsdk:"threads"`
	Cores       types.Int64  `tfsdk:"cores"`
	MemoryInMB  types.Int64  `tfsdk:"memory_in_mb"`
	StorageInGB types.Int64  `tfsdk:"storage_in_gb"`
	StorageType types.String `tfsdk:"storage_type"`
	GPUCount    types.Int64  `tfsdk:"gpu_count"`
	GPUType     types.String `tfsdk:"gpu_type"`
}

// ToTerraform converts a client.SKUResponse into a terraform Model.
func ToTerraform(sku client.SKUResponse) Model {
	return Model{
		ID:          types.StringValue(sku.ID),
		Name:        types.StringValue(sku.Name),
		Family:      types.StringValue(sku.Family),
		Threads:     types.Int64Value(int64(sku.Threads)),     // nolint:gosec
		Cores:       types.Int64Value(int64(sku.Cores)),       // nolint:gosec
		MemoryInMB:  types.Int64Value(int64(sku.MemoryInMB)),  // nolint:gosec
		StorageInGB: types.Int64Value(int64(sku.StorageInGB)), // nolint:gosec
		StorageType: types.StringValue(sku.StorageType),
		GPUCount:    types.Int64Value(int64(sku.GPUCount)), // nolint:gosec
		GPUType:     types.StringValue(sku.GPUType),
	}
}

// DataSourceAttributes returns the data source schema attributes describing a SKU.
func DataSourceAttributes() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"id": dsschema.StringAttribute{
			Description: "The ID of the SKU.",
			Computed:    true,
		},
		"name": dsschema.StringAttribute{
			Description: "The name of the SKU.",
			Computed:    true,
		},
		"family": dsschema.StringAttribute{
			Description: "The family of the SKU.",
			Computed:    true,
		},
		"threads": dsschema.Int64Attribute{
			Description: "The number of threads.",
			Computed:    true,
		},
		"cores": dsschema.Int64Attribute{
			Description: "The number of cores.",
			Computed:    true,
		},
		"memory_in_mb": dsschema.Int64Attribute{
			Description: "The amount of memory in MB.",
			Computed:    true,
		},
		"storage_in_gb": dsschema.Int64Attribute{
			Description: "The amount of storage in GB.",
			Computed:    true,
		},
		"storage_type": dsschema.StringAttribute{
			Description: "The type of storage.",
			Computed:    true,
		},
		"gpu_count": dsschema.Int64Attribute{
			Description: "The number of GPUs.",
			Computed:    true,
		},
		"gpu_type": dsschema.StringAttribute{
			Description: "The type of GPU.",
			Computed:    true,
		},
	}
}

// ResourceAttributes returns the resource schema attributes describing a SKU.
func ResourceAttributes() map[string]rsschema.Attribute {
	return map[string]rsschema.Attribute{
		"id": rsschema.StringAttribute{
			Description: "The ID of the SKU.",
			Computed:    true,
		},
		"name": rsschema.StringAttribute{
			Description: "The name of the SKU.",
			Computed:    true,
		},
		"family": rsschema.StringAttribute{
			Description: "The family of the SKU.",
			Computed:    true,
		},
		"threads": rsschema.Int64Attribute{
			Description: "The number of threads.",
			Computed:    true,
		},
		"cores": rsschema.Int64Attribute{
			Description: "The number of cores.",
			Computed:    true,
		},
		"memory_in_mb": rsschema.Int64Attribute{
			Description: "The amount of memory in MB.",
			Computed:    true,
		},
		"storage_in_gb": rsschema.Int64Attribute{
			Description: "The amount of storage in GB.",
			Computed:    true,
		},
		"storage_type": rsschema.StringAttribute{
			Description: "The type of storage.",
			Computed:    true,
		},
		"gpu_count": rsschema.Int64Attribute{
			Description: "The number of GPUs.",
			Computed:    true,
		},
		"gpu_type": rsschema.StringAttribute{
			Description: "The type of GPU.",
			Computed:    true,
		},
	}
}
