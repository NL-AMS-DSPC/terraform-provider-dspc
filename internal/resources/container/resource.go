// Package container implements the Terraform resource for managing container deployments.
package container

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing container resources.
type ResourceClient interface {
	CreateDeployment(ctx context.Context, req client.Container) (*client.Container, error)
	GetDeployment(ctx context.Context, name string) (*client.Container, error)
	DeleteDeployment(ctx context.Context, name string) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	TenantID     types.String `tfsdk:"tenant_id"`
	Image        types.String `tfsdk:"image"`
	SkuID        types.String `tfsdk:"sku_id"`
	Port         types.Int64  `tfsdk:"port"`
	Command      types.String `tfsdk:"command"`
	Args         types.List   `tfsdk:"args"`
	Env          types.List   `tfsdk:"env"`
	WorkingDir   types.String `tfsdk:"working_dir"`
	User         types.String `tfsdk:"user"`
	Group        types.String `tfsdk:"group"`
	Replicas     types.Int64  `tfsdk:"replicas"`
	Tags         types.Map    `tfsdk:"tags"`
	RegistryAuth types.Object `tfsdk:"registry_auth"`
	Secrets      types.List   `tfsdk:"secrets"`
}

// registryAuthModel mirrors the registry_auth nested attribute for ObjectAs extraction.
type registryAuthModel struct {
	Server   types.String `tfsdk:"server"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// secretModel mirrors one secrets[] entry for ElementsAs extraction.
type secretModel struct {
	EnvName types.String `tfsdk:"env_name"`
	Value   types.String `tfsdk:"value"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a container deployment in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the container deployment.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the container deployment. Must be unique within the tenant.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Description: "The identifier of the tenant that owns the container deployment.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "The container image to deploy (e.g. \"nginx:latest\").",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sku_id": schema.StringAttribute{
				Description: "The SKU id sizing the deployment (e.g. \"gp-1\"). List available SKUs via GET /api/containers/v1/skus.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The port exposed by the container. Must be greater than 0.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"command": schema.StringAttribute{
				Description: "The command to run in the container.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"args": schema.ListAttribute{
				Description: "Arguments to pass to the container command.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"env": schema.ListAttribute{
				Description: "Environment variables for the container (e.g. [\"KEY=value\"]).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"working_dir": schema.StringAttribute{
				Description: "The working directory inside the container.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Description: "The user to run the container as.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group": schema.StringAttribute{
				Description: "The group to run the container as.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of replicas. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				Description: "Tags to attach to the container deployment.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"registry_auth": schema.SingleNestedAttribute{
				Description: "Private registry pull credentials. Write-only: never returned by the API on read.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Description: "Registry server hostname (e.g. \"harbor.example.com\").",
						Required:    true,
					},
					"username": schema.StringAttribute{
						Description: "Registry username.",
						Required:    true,
					},
					"password": schema.StringAttribute{
						Description: "Registry password. Write-only: never stored in Terraform state (Terraform >= 1.11).",
						Required:    true,
						Sensitive:   true,
						WriteOnly:   true,
					},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Description: "Runtime secrets exposed as environment variables. Write-only: never returned by the API on read.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"env_name": schema.StringAttribute{
							Description: "Environment variable name to set inside the container.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Secret value. Write-only: never stored in Terraform state (Terraform >= 1.11).",
							Required:    true,
							Sensitive:   true,
							WriteOnly:   true,
						},
					},
				},
			},
		},
	}
}

// Configure stores the API client for the resource to use.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	ascClient, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if ascClient.Containers == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected container service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = ascClient.Containers
}

// Create creates a new container deployment.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// WriteOnly attribute values are null in plan/state — must be read from config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.Container{
		Name:  plan.Name.ValueString(),
		Image: plan.Image.ValueString(),
		SkuID: plan.SkuID.ValueString(),
		Port:  safeInt32(plan.Port.ValueInt64()),
	}

	if !plan.Command.IsNull() && !plan.Command.IsUnknown() {
		createReq.Command = plan.Command.ValueString()
	}

	if !plan.WorkingDir.IsNull() && !plan.WorkingDir.IsUnknown() {
		createReq.WorkingDir = plan.WorkingDir.ValueString()
	}

	if !plan.User.IsNull() && !plan.User.IsUnknown() {
		createReq.User = plan.User.ValueString()
	}

	if !plan.Group.IsNull() && !plan.Group.IsUnknown() {
		createReq.Group = plan.Group.ValueString()
	}

	if !plan.Replicas.IsNull() && !plan.Replicas.IsUnknown() {
		createReq.Replicas = safeInt32(plan.Replicas.ValueInt64())
	}

	if !plan.Args.IsNull() && !plan.Args.IsUnknown() {
		var args []string
		resp.Diagnostics.Append(plan.Args.ElementsAs(ctx, &args, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Args = args
	}

	if !plan.Env.IsNull() && !plan.Env.IsUnknown() {
		var env []string
		resp.Diagnostics.Append(plan.Env.ElementsAs(ctx, &env, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Env = env
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tagsMap map[string]string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range tagsMap {
			createReq.Tags = append(createReq.Tags, client.ContainerTag{Key: k, Value: v})
		}
	}

	if !config.RegistryAuth.IsNull() && !config.RegistryAuth.IsUnknown() {
		var ra registryAuthModel
		resp.Diagnostics.Append(config.RegistryAuth.As(ctx, &ra, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.RegistryAuth = &client.RegistryAuth{
			Server:   ra.Server.ValueString(),
			Username: ra.Username.ValueString(),
			Password: ra.Password.ValueString(),
		}
	}

	if !config.Secrets.IsNull() && !config.Secrets.IsUnknown() {
		var secrets []secretModel
		resp.Diagnostics.Append(config.Secrets.ElementsAs(ctx, &secrets, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, s := range secrets {
			createReq.Secrets = append(createReq.Secrets, client.RuntimeSecret{
				EnvName: s.EnvName.ValueString(),
				Value:   s.Value.ValueString(),
			})
		}
	}

	container, err := r.client.CreateDeployment(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating container",
			fmt.Sprintf("Could not create container deployment: %s", err.Error()),
		)
		return
	}

	mapStateFromContainer(ctx, &plan, container, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the container deployment data from the API.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	container, err := r.client.GetDeployment(ctx, state.Name.ValueString())
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading container",
			fmt.Sprintf("Could not read container deployment: %s", err.Error()),
		)
		return
	}

	// Preserve write-only fields — API never returns them; prior state is the only source.
	priorRegistryAuth := state.RegistryAuth
	priorSecrets := state.Secrets

	mapStateFromContainer(ctx, &state, container, &resp.Diagnostics)

	state.RegistryAuth = priorRegistryAuth
	state.Secrets = priorSecrets

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for container deployments at this time.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Container deployment updates are not supported. Changes require recreation.",
	)
}

// Delete removes the container deployment.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDeployment(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting container",
			fmt.Sprintf("Could not delete container deployment: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the container deployment state.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// mapStateFromContainer maps a Container API response to the Terraform ResourceModel.
func mapStateFromContainer(ctx context.Context, model *ResourceModel, c *client.Container, diags *diag.Diagnostics) {
	model.ID = types.StringValue(c.Name)
	model.Name = types.StringValue(c.Name)
	model.TenantID = types.StringValue(c.TenantID)
	model.Image = types.StringValue(c.Image)
	// SkuID is write-only on the API — GET returns "". Preserve plan/state value.
	if c.SkuID != "" {
		model.SkuID = types.StringValue(c.SkuID)
	}
	model.Port = types.Int64Value(int64(c.Port))

	if c.Replicas > 0 {
		model.Replicas = types.Int64Value(int64(c.Replicas))
	}

	if c.Command != "" {
		model.Command = types.StringValue(c.Command)
	}
	if c.WorkingDir != "" {
		model.WorkingDir = types.StringValue(c.WorkingDir)
	}
	if c.User != "" {
		model.User = types.StringValue(c.User)
	}
	if c.Group != "" {
		model.Group = types.StringValue(c.Group)
	}

	if len(c.Args) > 0 {
		argsList, d := types.ListValueFrom(ctx, types.StringType, c.Args)
		diags.Append(d...)
		model.Args = argsList
	}

	if len(c.Env) > 0 {
		envList, d := types.ListValueFrom(ctx, types.StringType, c.Env)
		diags.Append(d...)
		model.Env = envList
	}

	if len(c.Tags) > 0 {
		tagsMap := make(map[string]string, len(c.Tags))
		for _, t := range c.Tags {
			tagsMap[t.Key] = t.Value
		}
		tagsValue, d := types.MapValueFrom(ctx, types.StringType, tagsMap)
		diags.Append(d...)
		model.Tags = tagsValue
	}
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (err.Error() == "resource not found" ||
		len(err.Error()) > 14 && err.Error()[:14] == "API error 404:")
}

// safeInt32 converts int64 to int32, clamping to math.MaxInt32 to prevent overflow.
func safeInt32(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(v) // #nosec G115 -- overflow is guarded by the check above
}
