// Package image implements the image data source.
package image

import (
	"context"
	"fmt"

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

// DataClient defines an interface for interacting with image data operations.
type DataClient interface {
	ListImages(ctx context.Context) ([]client.ImageResponse, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	Images []Model `tfsdk:"images"`
}

// Model represents an OS image.
type Model struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Family                 types.String `tfsdk:"family"`
	Distribution           types.String `tfsdk:"distribution"`
	Release                types.String `tfsdk:"release"`
	RequiresLicense        types.Bool   `tfsdk:"requires_license"`
	LicenseInfo            types.String `tfsdk:"license_info"`
	SupportedArchitectures types.List   `tfsdk:"supported_architectures"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of the image.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of the image.",
			Computed:    true,
		},
		"family": schema.StringAttribute{
			Description: "The family of the image.",
			Computed:    true,
		},
		"distribution": schema.StringAttribute{
			Description: "The distribution of the image.",
			Computed:    true,
		},
		"release": schema.StringAttribute{
			Description: "The release of the image.",
			Computed:    true,
		},
		"requires_license": schema.BoolAttribute{
			Description: "Whether the image requires a license.",
			Computed:    true,
		},
		"license_info": schema.StringAttribute{
			Description: "License information for the image.",
			Computed:    true,
		},
		"supported_architectures": schema.ListAttribute{
			Description: "The list of architectures supported by the image.",
			Computed:    true,
			ElementType: types.StringType,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all available OS images in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"images": schema.ListNestedAttribute{
				Description: "List of available OS images.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: attrs,
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

	dataClient, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Images == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected images service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Images
}

// Read reads the data from the API and stores it in the state.
func (d *DataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DataSourceModel

	images, err := d.client.ListImages(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing images",
			fmt.Sprintf("Could not list images: %s", err.Error()),
		)
		return
	}

	state.Images = make([]Model, len(images))
	for i, img := range images {
		model := ToTerraform(ctx, img, &resp.Diagnostics)
		state.Images[i] = model
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ToTerraform converts a client.ImageResponse into a terraform Model.
func ToTerraform(ctx context.Context, img client.ImageResponse, diags *diag.Diagnostics) Model {
	archs, d := types.ListValueFrom(ctx, types.StringType, img.SupportedArchitectures)
	diags.Append(d...)

	return Model{
		ID:                     types.StringValue(img.ID),
		Name:                   types.StringValue(img.Name),
		Family:                 types.StringValue(img.Family),
		Distribution:           types.StringValue(img.Distribution),
		Release:                types.StringValue(img.Release),
		RequiresLicense:        types.BoolValue(img.RequiresLicense),
		LicenseInfo:            types.StringValue(img.LicenseInfo),
		SupportedArchitectures: archs,
	}
}
