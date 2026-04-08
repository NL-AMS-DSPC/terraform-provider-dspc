package postgresql

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

// ResourceClient defines the interface for the methods needed to manage PostgreSQL instances.
type ResourceClient interface {
	CreatePostgreSQLInstance(ctx context.Context, req client.CreatePostgreSQLInstanceRequest) (*client.PostgreSQLInstance, error)
	GetPostgreSQLInstance(ctx context.Context, instanceName string) (*client.PostgreSQLInstance, error)
	ListPostgreSQLInstances(ctx context.Context) (*client.ListPostgreSQLInstancesResponse, error)
	UpdatePostgreSQLInstance(ctx context.Context, instanceName string, req client.UpdatePostgreSQLInstanceRequest) (*client.PostgreSQLInstance, error)
	DeletePostgreSQLInstance(ctx context.Context, instanceName string) error
}

// Resource implements the Terraform resource for managing PostgreSQL instances in the DSPC platform.
type Resource struct {
	client ResourceClient
}

// TagModel represents a key-value pair for tagging resources.
type TagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// ResourceModel represents the schema for the PostgreSQL resource in Terraform.
type ResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Size    types.String `tfsdk:"size"`
	Version types.String `tfsdk:"version"`
	VPC     types.String `tfsdk:"vpc"`
	Tags    []TagModel   `tfsdk:"tags"`
}

// NewResource creates a new instance of the PostgreSQL resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata returns the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql"
}

// Schema defines the schema for the PostgreSQL resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PostgreSQL instance in the DSPC platform.",
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
			"size": schema.StringAttribute{
				Required:    true,
				Description: "Size of the database storage, e.g. 500Mi, 1Gi.",
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Version of the database engine. One of: POSTGRES_15, POSTGRES_16, POSTGRES_17, POSTGRES_18.",
			},
			"vpc": schema.StringAttribute{
				Required:    true,
				Description: "Name of the VPC network where this database should be added to.",
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

	if dataClient.Network == nil {
		resp.Diagnostics.AddError(
			"Unexpected resource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Network
}

// Create handles the creation of a new PostgreSQL instance based on the provided configuration.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreatePostgreSQLInstanceRequest{
		Name:    plan.Name.ValueString(),
		Size:    plan.Size.ValueString(),
		Version: client.DatabaseVersion(plan.Version.ValueString()),
		VPC:     plan.VPC.ValueString(),
		Tags:    toClientTags(plan.Tags),
	}

	instance, err := r.client.CreatePostgreSQLInstance(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating PostgreSQL instance",
			fmt.Sprintf("An error was encountered when creating the PostgreSQL instance: %s", err.Error()),
		)
		return
	}

	plan = toResourceModel(instance)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read retrieves the current state of the PostgreSQL instance and updates the Terraform state accordingly.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instance, err := r.client.GetPostgreSQLInstance(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PostgreSQL instance",
			fmt.Sprintf("An error was encountered when reading the PostgreSQL instance: %s", err.Error()),
		)
		return
	}

	state = toResourceModel(instance)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update applies any in-place changes to the PostgreSQL instance (size, version, vpc, tags).
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdatePostgreSQLInstanceRequest{
		Name:    plan.Name.ValueString(),
		Size:    plan.Size.ValueString(),
		Version: client.DatabaseVersion(plan.Version.ValueString()),
		VPC:     plan.VPC.ValueString(),
		Tags:    toClientTags(plan.Tags),
	}

	instance, err := r.client.UpdatePostgreSQLInstance(ctx, plan.Name.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating PostgreSQL instance",
			fmt.Sprintf("An error was encountered when updating the PostgreSQL instance: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, toResourceModel(instance))...)
}

// Delete removes the PostgreSQL instance from the DSPC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePostgreSQLInstance(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting PostgreSQL instance",
			fmt.Sprintf("An error was encountered when deleting the PostgreSQL instance: %s", err.Error()),
		)
	}
}

// ImportState allows users to import existing PostgreSQL instances into Terraform state using the instance name as the identifier.
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

func toResourceModel(instance *client.PostgreSQLInstance) ResourceModel {
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
