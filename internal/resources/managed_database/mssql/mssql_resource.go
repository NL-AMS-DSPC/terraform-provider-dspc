package mssql

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for the methods needed to manage MSSQL instances.
type ResourceClient interface {
	CreateMSSQLInstance(ctx context.Context, req client.CreateMSSQLInstanceRequest) (*client.MSSQLInstance, error)
	GetMSSQLInstance(ctx context.Context, instanceName string) (*client.MSSQLInstance, error)
	ListMSSQLInstances(ctx context.Context) (*client.ListMSSQLInstancesResponse, error)
	UpdateMSSQLInstance(ctx context.Context, instanceName string, req client.UpdateMSSQLInstanceRequest) (*client.MSSQLInstance, error)
	DeleteMSSQLInstance(ctx context.Context, instanceName string) error
}

// Resource implements the Terraform resource for managing MSSQL instances in the DSPC platform.
type Resource struct {
	client ResourceClient
}

// TagModel represents a key-value pair for tagging resources.
type TagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// AdditionalConfiguration for a mssql database
type AdditionalConfiguration struct {
	LicenseKey types.String `tfsdk:"license_key"`
}

// ResourceModel represents the schema for the MSSQL resource in Terraform.
type ResourceModel struct {
	Name                    types.String            `tfsdk:"name"`
	SkuSize                 types.String            `tfsdk:"sku_size"`
	Version                 types.String            `tfsdk:"version"`
	VPCID                   types.String            `tfsdk:"vpc_id"`
	Tags                    []TagModel              `tfsdk:"tags"`
	AdditionalConfiguration AdditionalConfiguration `tfsdk:"additional_configuration"`
}

// NewResource creates a new instance of the MSSQL resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mssql"
}

// Schema defines the schema for the MSSQL resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Microsoft SQL Server instance in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name for the database instance. Must be 1-63 lowercase alphanumeric characters or hyphens, and must start and end with an alphanumeric character.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`),
						"must be 1-63 lowercase alphanumeric characters or hyphens, and must start and end with an alphanumeric character",
					),
					stringvalidator.LengthBetween(1, 63),
				},
			},
			"sku_size": schema.StringAttribute{
				Required:    true,
				Description: "Sku size per instance node, e.g. gp-2, gp-4, etc",
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Version of the database engine. One of: MSSQL_2025_17, MSSQL_2022_16, MSSQL_2019_15, MSSQL_2017_14.",
			},
			"vpc_id": schema.StringAttribute{
				Required:    true,
				Description: "GUID of the VPC network where this database should be added to.",
			},
			"additional_configuration": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Additional configuration options for the MSSQL instance.",
				Attributes: map[string]schema.Attribute{
					"license_key": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "License key for the MSSQL instance.",
					},
				},
			},
			"tags": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Tags to apply to the database instance.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "Tag key. Must be a qualified name: an optional DNS subdomain prefix (max 253 chars) followed by a slash and a name (max 63 chars), or just a name. Name must start and end with alphanumeric characters and may contain alphanumeric characters, hyphens, underscores, and dots.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/)?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$`),
									"must be a qualified name: an optional DNS subdomain prefix followed by '/' and a name segment, or just a name segment. Name segments must start and end with alphanumeric characters and may contain alphanumeric characters, hyphens, underscores, and dots.",
								),
								stringvalidator.LengthAtMost(316),
							},
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "Tag value. Must be 63 characters or less, start and end with alphanumeric characters, and may contain alphanumeric characters, hyphens, underscores, and dots. May also be empty.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$|^$`),
									"must start and end with alphanumeric characters and may only contain alphanumeric characters, hyphens, underscores, and dots, or be empty",
								),
								stringvalidator.LengthAtMost(63),
							},
						},
					},
				},
			},
		},
	}
}

// Configure sets the client for the resource based on the provider configuration.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.ManagedDB == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected managed database service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.ManagedDB
}

// Create handles the creation of a new MSSQL instance based on the provided configuration.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateMSSQLInstanceRequest{
		Name:                     plan.Name.ValueString(),
		SkuSize:                  plan.SkuSize.ValueString(),
		Version:                  client.DatabaseVersion(plan.Version.ValueString()),
		VPCID:                    plan.VPCID.ValueString(),
		Tags:                     toClientTags(plan.Tags),
		AdditionalConfigurations: toClientAdditionalConfig(plan.AdditionalConfiguration),
	}

	instance, err := r.client.CreateMSSQLInstance(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating MSSQL instance",
			fmt.Sprintf("An error was encountered when creating the MSSQL instance: %s", err.Error()),
		)
		return
	}

	newState := toResourceModel(instance)
	newState.AdditionalConfiguration = plan.AdditionalConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Read retrieves the current state of the MSSQL instance and updates the Terraform state accordingly.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.client.GetMSSQLInstance(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MSSQL instance",
			fmt.Sprintf("An error was encountered when reading the MSSQL instance: %s", err.Error()),
		)
		return
	}

	newState := toResourceModel(instance)
	newState.AdditionalConfiguration = state.AdditionalConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Update applies any in-place changes to the MSSQL instance (size, version, vpc, tags).
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateMSSQLInstanceRequest{
		Name:                     plan.Name.ValueString(),
		SkuSize:                  plan.SkuSize.ValueString(),
		Version:                  client.DatabaseVersion(plan.Version.ValueString()),
		VPCID:                    plan.VPCID.ValueString(),
		Tags:                     toClientTags(plan.Tags),
		AdditionalConfigurations: toClientAdditionalConfig(plan.AdditionalConfiguration),
	}

	instance, err := r.client.UpdateMSSQLInstance(ctx, plan.Name.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating MSSQL instance",
			fmt.Sprintf("An error was encountered when updating the MSSQL instance: %s", err.Error()),
		)
		return
	}

	newState := toResourceModel(instance)
	newState.AdditionalConfiguration = plan.AdditionalConfiguration

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// Delete removes the MSSQL instance from the DSPC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMSSQLInstance(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting MSSQL instance",
			fmt.Sprintf("An error was encountered when deleting the MSSQL instance: %s", err.Error()),
		)
	}
}

// ImportState allows users to import existing MSSQL instances into Terraform state using the instance name as the identifier.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func toClientAdditionalConfig(cfg AdditionalConfiguration) map[string]any {
	result := make(map[string]any)
	if !cfg.LicenseKey.IsNull() && !cfg.LicenseKey.IsUnknown() {
		result["license_key"] = cfg.LicenseKey.ValueString()
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func toClientTags(tags []TagModel) []client.Tag {
	if tags == nil {
		return nil
	}
	result := make([]client.Tag, len(tags))
	for i, t := range tags {
		result[i] = client.Tag{
			Key:   t.Key.ValueString(),
			Value: t.Value.ValueString(),
		}
	}
	return result
}

func toResourceModel(instance *client.MSSQLInstance) ResourceModel {
	model := ResourceModel{
		Name:    types.StringValue(instance.Name),
		SkuSize: types.StringValue(instance.SkuSize),
		Version: types.StringValue(string(instance.Version)),
		VPCID:   types.StringValue(instance.VPCID),
	}
	if len(instance.Tags) > 0 {
		model.Tags = make([]TagModel, len(instance.Tags))
		for i, t := range instance.Tags {
			model.Tags[i] = TagModel{
				Key:   types.StringValue(t.Key),
				Value: types.StringValue(t.Value),
			}
		}
	}
	return model
}
