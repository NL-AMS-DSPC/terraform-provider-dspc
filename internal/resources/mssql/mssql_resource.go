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

type ResourceClient interface {
	CreateMSSQLInstance(ctx context.Context, req client.CreateMSSQLInstanceRequest) (*client.MSSQLInstance, error)
	GetMSSQLInstance(ctx context.Context, instanceName string) (*client.MSSQLInstance, error)
	ListMSSQLInstances(ctx context.Context) (*client.ListMSSQLInstancesResponse, error)
}

type Resource struct {
	client ResourceClient
}

type TagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type ResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Size    types.String `tfsdk:"size"`
	Version types.String `tfsdk:"version"`
	VPC     types.String `tfsdk:"vpc"`
	Tags    []TagModel   `tfsdk:"tags"`
}

func NewResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mssql"
}

func (r *Resource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
						regexp.MustCompile(`(?=\A[-a-z0-9]{1,63}\Z)\A[a-z0-9]+(-[a-z0-9]+)*\Z`),
						"must be 1-63 lowercase alphanumeric characters or hyphens, and must start and end with an alphanumeric character",
					),
				},
			},
			"size": schema.StringAttribute{
				Required:    true,
				Description: "Size of the database storage, e.g. 500Mi, 1Gi.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Version of the database engine. One of: MSSQL_2025_17, MSSQL_2022_16, MSSQL_2019_15, MSSQL_2017_14.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC network where this database should be added to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Tags to apply to the database instance.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "Tag key.",
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "Tag value.",
						},
					},
				},
			},
		},
	}
}

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

	if dataClient.Network == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Network
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateMSSQLInstanceRequest{
		Name:    plan.Name.ValueString(),
		Size:    plan.Size.ValueString(),
		Version: client.DatabaseVersion(plan.Version.ValueString()),
		VPC:     plan.VPC.ValueString(),
		Tags:    toClientTags(plan.Tags),
	}

	instance, err := r.client.CreateMSSQLInstance(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating MSSQL instance",
			fmt.Sprintf("An error was encountered when creating the MSSQL instance: %s", err.Error()),
		)
		return
	}

	plan = toResourceModel(instance)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

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

	state = toResourceModel(instance)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(_ context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating MSSQL instances is not currently supported. Please recreate the resource with the desired configuration.",
	)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"Deleting MSSQL instances is not currently supported via this provider.",
	)
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
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
		Size:    types.StringValue(instance.Size),
		Version: types.StringValue(string(instance.Version)),
		VPC:     types.StringValue(instance.VPC),
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
