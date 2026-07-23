package filestorage

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

type fsDataClient interface {
	GetFileStorage(ctx context.Context, name string) (*client.FileStorage, error)
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Size         types.String `tfsdk:"size"`
	Status       types.String `tfsdk:"status"`
	NFSMountPath types.String `tfsdk:"nfs_mount_path"`
}

// DataSource defines the data source implementation.
type DataSource struct {
	client fsDataClient
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage"
}

// Schema returns the schema for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing file storage volume from the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the file storage (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the file storage.",
				Required:    true,
			},
			"size": schema.StringAttribute{
				Description: "Size of the file storage (e.g. 100Gi).",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the file storage.",
				Computed:    true,
			},
			"nfs_mount_path": schema.StringAttribute{
				Description: "NFS path used to mount the file storage.",
				Computed:    true,
			},
		},
	}
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

	if c.FileStorage == nil {
		resp.Diagnostics.AddError(
			"Unexpected datasource configuration error",
			"Expected file storage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = c.FileStorage
}

// Read reads the file storage from the API and stores it in state.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fs, err := d.client.GetFileStorage(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file storage",
			fmt.Sprintf("Could not read file storage: %s", err.Error()),
		)
		return
	}

	state := DataSourceModel{
		ID:           types.StringValue(fs.Name),
		Name:         types.StringValue(fs.Name),
		Size:         types.StringValue(fs.Size),
		Status:       types.StringValue(fs.Status),
		NFSMountPath: types.StringValue(fs.NFSMountPath),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
