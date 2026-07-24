package securitygroupattachment

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
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines the interface for reading security group attachment data.
type DataClient interface {
	GetSecurityGroupAttachment(ctx context.Context, sgName, attachmentName string) (*client.SecurityGroupAttachment, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	AttachmentName    types.String `tfsdk:"attachment_name"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group_attachment"
}

// Schema updates the data source schema with the attributes.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a specific Security Group attachment in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the security group attachment.",
				Computed:    true,
			},
			"security_group_name": schema.StringAttribute{
				Description: "The name of the Security Group.",
				Required:    true,
			},
			"attachment_name": schema.StringAttribute{
				Description: "The name of the attachment resource.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sga, err := d.client.GetSecurityGroupAttachment(
		ctx,
		config.SecurityGroupName.ValueString(),
		config.AttachmentName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Security Group Attachment",
			fmt.Sprintf("Could not read security group attachment: %s", err.Error()),
		)
		return
	}

	var state DataSourceModel
	state.ID = types.StringValue(createStateID(config.SecurityGroupName.ValueString(), sga.Name))
	state.SecurityGroupName = types.StringValue(sga.SGRef)
	state.AttachmentName = types.StringValue(sga.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
