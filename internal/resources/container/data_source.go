// Package container provides Terraform resources and data sources for managing Containers.
package container

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines an interface for interacting with Container data operations.
type DataClient interface {
	GetDeployment(ctx context.Context, name string) (*client.Container, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Image      types.String `tfsdk:"image"`
	Port       types.Int32  `tfsdk:"port"`
	Command    types.String `tfsdk:"command"`
	Args       types.List   `tfsdk:"args"`
	Env        types.List   `tfsdk:"env"`
	WorkingDir types.String `tfsdk:"working_dir"`
	User       types.String `tfsdk:"user"`
	Group      types.String `tfsdk:"group"`
	Replicas   types.Int32  `tfsdk:"replicas"`
	Tags       types.List   `tfsdk:"tags"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a single container deployment in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the container deployment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the container deployment.",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "The image used by the container.",
				Computed:    true,
			},
			"port": schema.Int32Attribute{
				Description: "The port exposed by the container.",
				Computed:    true,
			},
			"command": schema.StringAttribute{
				Description: "The command to run in the container.",
				Computed:    true,
			},
			"args": schema.ListAttribute{
				Description: "Arguments passed to the container command.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"env": schema.ListAttribute{
				Description: "Environment variables for the container (e.g. [\"KEY=value\"]).",
				Computed:    true,
				ElementType: types.StringType,
			},
			"working_dir": schema.StringAttribute{
				Description: "The working directory inside the container.",
				Computed:    true,
			},
			"user": schema.StringAttribute{
				Description: "The user that the container is running as.",
				Computed:    true,
			},
			"group": schema.StringAttribute{
				Description: "The group that the container is running as.",
				Computed:    true,
			},
			"replicas": schema.Int32Attribute{
				Description: "The number of replicas.",
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags to attach to the container deployment.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"key":   types.StringType,
						"value": types.StringType,
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *DataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
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

	if dataClient.Containers == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Containers
}

// Read reads the data from the API and stores it in the state.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	container, err := d.client.GetDeployment(ctx, model.Name.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading container deployment",
			fmt.Sprintf("Could not get containers deployment: %s", err.Error()),
		)
		return
	}

	model.ID = types.StringValue(container.ID)
	model.Image = types.StringValue(container.Image)
	model.Port = types.Int32Value(container.Port)
	model.Command = types.StringValue(container.Command)
	model.WorkingDir = types.StringValue(container.WorkingDir)
	model.User = types.StringValue(container.User)
	model.Group = types.StringValue(container.Group)
	model.Replicas = types.Int32Value(container.Replicas)

	var diags diag.Diagnostics
	model.Args, diags = types.ListValueFrom(ctx, types.StringType, container.Args)
	resp.Diagnostics.Append(diags...)

	model.Env, diags = types.ListValueFrom(ctx, types.StringType, container.Env)
	resp.Diagnostics.Append(diags...)

	tags := make([]attr.Value, len(container.Tags))
	for i, tag := range container.Tags {
		tags[i], diags = types.ObjectValue(
			map[string]attr.Type{"key": types.StringType, "value": types.StringType},
			map[string]attr.Value{"key": types.StringValue(tag.Key), "value": types.StringValue(tag.Value)},
		)
		resp.Diagnostics.Append(diags...)
	}
	model.Tags, diags = types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key":   types.StringType,
			"value": types.StringType,
		},
	}, tags)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
