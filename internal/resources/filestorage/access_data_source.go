package filestorage

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
	_ datasource.DataSource              = &AccessDataSource{}
	_ datasource.DataSourceWithConfigure = &AccessDataSource{}
)

type accessDataClient interface {
	GetAccess(ctx context.Context, fileStorageName, targetType, targetName string) (*client.FileStorageAccess, error)
}

// AccessDataSourceModel describes the access data source data model.
type AccessDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	FileStorageName types.String `tfsdk:"file_storage_name"`
	TargetType      types.String `tfsdk:"target_type"`
	TargetName      types.String `tfsdk:"target_name"`
}

// AccessDataSource defines the data source implementation.
type AccessDataSource struct {
	client accessDataClient
}

// NewAccessDataSource creates a new AccessDataSource.
func NewAccessDataSource() datasource.DataSource {
	return &AccessDataSource{}
}

// Metadata updates the data source type name.
func (d *AccessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_storage_access"
}

// Schema returns the schema for the data source.
func (d *AccessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing file storage access entry from the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this access entry ({file_storage_name}:{target_type}:{target_name}).",
				Computed:    true,
			},
			"file_storage_name": schema.StringAttribute{
				Description: "Name of the file storage.",
				Required:    true,
			},
			"target_type": schema.StringAttribute{
				Description: "Type of the workload. Valid values: VirtualMachine, Container.",
				Required:    true,
			},
			"target_name": schema.StringAttribute{
				Description: "Name of the workload.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider-configured client to the data source.
func (d *AccessDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read reads the access entry from the API and stores it in state.
func (d *AccessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AccessDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	access, err := d.client.GetAccess(
		ctx,
		config.FileStorageName.ValueString(),
		config.TargetType.ValueString(),
		config.TargetName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file storage access",
			fmt.Sprintf("Could not read file storage access: %s", err.Error()),
		)
		return
	}

	state := AccessDataSourceModel{
		ID:              types.StringValue(accessStateID(access.FileStorageName, access.TargetType, access.TargetName)),
		FileStorageName: types.StringValue(access.FileStorageName),
		TargetType:      types.StringValue(access.TargetType),
		TargetName:      types.StringValue(access.TargetName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
