// Package objectstorage contains terraform definitions for the object storage resource
package objectstorage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

type objectStorageDataClient interface {
	GetBucket(ctx context.Context, id string) (*client.ObjectStorage, error)
}

type quotaDataModel struct {
	MaxSize types.String `tfsdk:"max_size"`
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	TenantID      types.String   `tfsdk:"tenant_id"`
	ReclaimPolicy types.String   `tfsdk:"reclaim_policy"`
	Endpoint      types.String   `tfsdk:"endpoint"`
	Region        types.String   `tfsdk:"region"`
	Quota         quotaDataModel `tfsdk:"quota"`
	Tags          []tagModel     `tfsdk:"tags"`
}

// DataSource defines the data source implementation.
type DataSource struct {
	client objectStorageDataClient
}

var objectStorageDataSchema = schema.Schema{
	Description: "Retrieves an existing object storage from the ASC platform.",
	Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required:    true,
			Description: "Unique identifier of the object storage.",
		},
		"name": schema.StringAttribute{
			Description: "Name of the object storage.",
			Computed:    true,
		},
		"tenant_id": schema.StringAttribute{
			Description: "Identifier of the tenant that owns the object storage.",
			Computed:    true,
		},
		"reclaim_policy": schema.StringAttribute{
			Description: "Reclaim policy of the object storage.",
			Computed:    true,
		},
		"endpoint": schema.StringAttribute{
			Description: "Endpoint of the object storage.",
			Computed:    true,
		},
		"region": schema.StringAttribute{
			Description: "Region of the object storage.",
			Computed:    true,
		},
		"tags": schema.ListNestedAttribute{
			Computed:    true,
			Description: "Tags applied to the database instance.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"key": schema.StringAttribute{
						Computed:    true,
						Description: "Tag key.",
					},
					"value": schema.StringAttribute{
						Computed:    true,
						Description: "Tag value.",
					},
				},
			},
		},
	},
	Blocks: map[string]schema.Block{
		"quota": schema.SingleNestedBlock{
			Description: "the quota configuration for the object storage",
			Attributes: map[string]schema.Attribute{
				"max_size": schema.StringAttribute{
					Computed:    true,
					Description: "the max size of the object storage",
				},
			},
		},
	},
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage"
}

// Schema returns the schema for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = objectStorageDataSchema
}

// Configure adds the provider-configured client to the data source.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if c.ObjectStorage == nil {
		resp.Diagnostics.AddError(
			"Unexpected datasource configuration error",
			"Expected object storage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = c.ObjectStorage
}

// Read reads the object storage from the API and stores it in state.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objectStorage, err := d.client.GetBucket(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading object storage",
			fmt.Sprintf("Could not read object storage: %s", err.Error()),
		)
		return
	}

	state := DataSourceModel{
		ID:            types.StringValue(objectStorage.ID),
		Name:          types.StringValue(objectStorage.Name),
		TenantID:      types.StringValue(objectStorage.TenantID),
		ReclaimPolicy: types.StringValue(objectStorage.ReclaimPolicy),
		Endpoint:      types.StringValue(objectStorage.Endpoint),
		Region:        types.StringValue(objectStorage.Region),
		Quota: quotaDataModel{
			MaxSize: types.StringValue(objectStorage.Quota.MaxSize),
		},
	}
	if len(objectStorage.Tags) > 0 {
		state.Tags = make([]tagModel, len(objectStorage.Tags))
		for i, t := range objectStorage.Tags {
			state.Tags[i] = tagModel{
				Key:   types.StringValue(t.Key),
				Value: types.StringValue(t.Value),
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
