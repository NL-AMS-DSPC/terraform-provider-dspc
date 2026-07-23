package function

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

// safeInt32Convert safely converts int64 to int32, clamping to int32 bounds if necessary
func safeInt32Convert(val int64) int32 {
	if val > 2147483647 {
		return 2147483647 // int32 max
	}
	if val < -2147483648 {
		return -2147483648 // int32 min
	}
	return int32(val)
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing function resources.
// It provides methods to create, delete, retrieve, and list functions.
type ResourceClient interface {
	CreateFunction(ctx context.Context, req client.CreateFunctionRequest) (*client.Function, error)
	UpdateFunction(ctx context.Context, name string, req client.UpdateFunctionRequest) (*client.Function, error)
	DeleteFunction(ctx context.Context, name string) error
	GetFunction(ctx context.Context, name string) (*client.Function, error)
	ListFunctions(ctx context.Context) ([]*client.Function, error)
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// EnvVarModel represents an environment variable
type EnvVarModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// secretModel mirrors one secrets[] entry. Value is write-only: never returned by the API.
type secretModel struct {
	EnvName types.String `tfsdk:"env_name"`
	Value   types.String `tfsdk:"value"`
}

// registryAuthModel mirrors the registry_auth nested attribute. Write-only on the API.
type registryAuthModel struct {
	Server   types.String `tfsdk:"server"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// ResourcesModel represents resource limits and requests
type ResourcesModel struct {
	CPURequest    types.String `tfsdk:"cpu_request"`
	CPULimit      types.String `tfsdk:"cpu_limit"`
	MemoryRequest types.String `tfsdk:"memory_request"`
	MemoryLimit   types.String `tfsdk:"memory_limit"`
}

// ConcurrencyModel represents concurrency settings
type ConcurrencyModel struct {
	Limit types.Int64 `tfsdk:"limit"`
}

// ProbeModel represents health check probe settings
type ProbeModel struct {
	Path                types.String `tfsdk:"path"`
	Port                types.Int64  `tfsdk:"port"`
	InitialDelaySeconds types.Int64  `tfsdk:"initial_delay_seconds"`
	PeriodSeconds       types.Int64  `tfsdk:"period_seconds"`
	TimeoutSeconds      types.Int64  `tfsdk:"timeout_seconds"`
	FailureThreshold    types.Int64  `tfsdk:"failure_threshold"`
}

// HealthChecksModel represents health check configuration
type HealthChecksModel struct {
	Liveness  *ProbeModel `tfsdk:"liveness"`
	Readiness *ProbeModel `tfsdk:"readiness"`
}

// TagModel represents a key-value tag
type TagModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

var tagObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	},
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID                  types.String       `tfsdk:"id"`
	TenantID            types.String       `tfsdk:"tenant_id"`
	Name                types.String       `tfsdk:"name"`
	Image               types.String       `tfsdk:"image"`
	Port                types.Int64        `tfsdk:"port"`
	Env                 []EnvVarModel      `tfsdk:"env"`
	Secrets             types.List         `tfsdk:"secrets"`
	RegistryAuth        types.Object       `tfsdk:"registry_auth"`
	Resources           *ResourcesModel    `tfsdk:"resources"`
	Concurrency         *ConcurrencyModel  `tfsdk:"concurrency"`
	HealthChecks        *HealthChecksModel `tfsdk:"health_checks"`
	Tags                types.Set          `tfsdk:"tags"`
	URL                 types.String       `tfsdk:"url"`
	Status              types.String       `tfsdk:"status"`
	LatestReadyRevision types.String       `tfsdk:"latest_ready_revision"`
	CreatedAt           types.String       `tfsdk:"created_at"`
	UpdatedAt           types.String       `tfsdk:"updated_at"`
}

func expandTagSet(ctx context.Context, tagSet types.Set) ([]client.Tag, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tagSet.IsNull() || tagSet.IsUnknown() {
		return nil, diags
	}

	var tagModels []TagModel
	diags.Append(tagSet.ElementsAs(ctx, &tagModels, false)...)
	if diags.HasError() {
		return nil, diags
	}

	tags := make([]client.Tag, len(tagModels))
	for i, tag := range tagModels {
		tags[i] = client.Tag{
			Key:   tag.Key.ValueString(),
			Value: tag.Value.ValueString(),
		}
	}

	return tags, diags
}

func flattenTags(ctx context.Context, tags []client.Tag) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tags) == 0 {
		return types.SetNull(tagObjectType), diags
	}

	tagModels := make([]TagModel, len(tags))
	for i, tag := range tags {
		tagModels[i] = TagModel{
			Key:   types.StringValue(tag.Key),
			Value: types.StringValue(tag.Value),
		}
	}

	tagSet, setDiags := types.SetValueFrom(ctx, tagObjectType, tagModels)
	diags.Append(setDiags...)

	return tagSet, diags
}

// expandWriteOnly reads the write-only secrets[] and registry_auth from the config model
// (their values are null in plan/state, so config is the only source) and returns the
// client representations to attach to a create/update request.
func expandWriteOnly(ctx context.Context, config ResourceModel) ([]client.RuntimeSecret, *client.RegistryAuth, diag.Diagnostics) {
	var diags diag.Diagnostics

	var secrets []client.RuntimeSecret
	if !config.Secrets.IsNull() && !config.Secrets.IsUnknown() {
		var models []secretModel
		diags.Append(config.Secrets.ElementsAs(ctx, &models, false)...)
		if diags.HasError() {
			return nil, nil, diags
		}
		for _, s := range models {
			secrets = append(secrets, client.RuntimeSecret{
				EnvName: s.EnvName.ValueString(),
				Value:   s.Value.ValueString(),
			})
		}
	}

	var registryAuth *client.RegistryAuth
	if !config.RegistryAuth.IsNull() && !config.RegistryAuth.IsUnknown() {
		var ra registryAuthModel
		diags.Append(config.RegistryAuth.As(ctx, &ra, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, nil, diags
		}
		registryAuth = &client.RegistryAuth{
			Server:   ra.Server.ValueString(),
			Username: ra.Username.ValueString(),
			Password: ra.Password.ValueString(),
		}
	}

	return secrets, registryAuth, diags
}

// NewFunctionResource creates a new Resource.
func NewFunctionResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a function in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the function.",
				Computed:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Identifier of the tenant that owns the function.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the function. Must be unique within the platform.",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "The container image for the function (e.g., 'gcr.io/knative-samples/helloworld-go').",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port the container listens on (1024-65535).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(8080),
			},
			"url": schema.StringAttribute{
				Description: "The URL of the function.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the function (e.g., \"pending\", \"ready\").",
				Computed:    true,
			},
			"latest_ready_revision": schema.StringAttribute{
				Description: "The latest ready revision of the function.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp of the function.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "The last update timestamp of the function.",
				Computed:    true,
			},
			"env": schema.ListNestedAttribute{
				Description: "Environment variables for the function.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the environment variable.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "The value of the environment variable.",
							Required:    true,
						},
					},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Description: "Runtime secrets exposed as environment variables. Write-only: never returned by the API on read. Changing them forces resource recreation.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"env_name": schema.StringAttribute{
							Description: "The environment variable name to set inside the function.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Secret value, stored in OpenBao. Write-only: never stored in Terraform state (Terraform >= 1.11).",
							Required:    true,
							Sensitive:   true,
							WriteOnly:   true,
						},
					},
				},
			},
			"registry_auth": schema.SingleNestedAttribute{
				Description: "Private registry pull credentials for the image. Write-only: never returned by the API on read. Changing them forces resource recreation.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Description: "Registry server hostname (e.g. \"harbor.example.com\"). Optional; derived from the image when omitted.",
						Optional:    true,
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
			"tags": schema.SetNestedAttribute{
				Description: "Tags for the function.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "The tag key.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "The tag value.",
							Required:    true,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"concurrency": schema.SingleNestedBlock{
				Description: "Concurrency configuration for the function.",
				Attributes: map[string]schema.Attribute{
					"limit": schema.Int64Attribute{
						Description: "Maximum number of concurrent requests (0 for unlimited).",
						Optional:    true,
					},
				},
			},
			"resources": schema.SingleNestedBlock{
				Description: "Resource limits and requests for the function.",
				Attributes: map[string]schema.Attribute{
					"cpu_request": schema.StringAttribute{
						Description: "CPU request (e.g., '100m').",
						Optional:    true,
					},
					"cpu_limit": schema.StringAttribute{
						Description: "CPU limit (e.g., '500m').",
						Optional:    true,
					},
					"memory_request": schema.StringAttribute{
						Description: "Memory request (e.g., '128Mi').",
						Optional:    true,
					},
					"memory_limit": schema.StringAttribute{
						Description: "Memory limit (e.g., '512Mi').",
						Optional:    true,
					},
				},
			},
			"health_checks": schema.SingleNestedBlock{
				Description: "Health check configuration for the function.",
				Blocks: map[string]schema.Block{
					"liveness": schema.SingleNestedBlock{
						Description: "Liveness probe configuration.",
						Attributes: map[string]schema.Attribute{
							"path": schema.StringAttribute{
								Description: "HTTP path for the probe.",
								Optional:    true,
							},
							"port": schema.Int64Attribute{
								Description: "Port for the probe.",
								Optional:    true,
							},
							"initial_delay_seconds": schema.Int64Attribute{
								Description: "Initial delay before probing starts.",
								Optional:    true,
							},
							"period_seconds": schema.Int64Attribute{
								Description: "How often to perform the probe.",
								Optional:    true,
							},
							"timeout_seconds": schema.Int64Attribute{
								Description: "Timeout for each probe attempt.",
								Optional:    true,
							},
							"failure_threshold": schema.Int64Attribute{
								Description: "Number of failures before marking as unhealthy.",
								Optional:    true,
							},
						},
					},
					"readiness": schema.SingleNestedBlock{
						Description: "Readiness probe configuration.",
						Attributes: map[string]schema.Attribute{
							"path": schema.StringAttribute{
								Description: "HTTP path for the probe.",
								Optional:    true,
							},
							"port": schema.Int64Attribute{
								Description: "Port for the probe.",
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(8080),
							},
							"initial_delay_seconds": schema.Int64Attribute{
								Description: "Initial delay before probing starts.",
								Optional:    true,
							},
							"period_seconds": schema.Int64Attribute{
								Description: "How often to perform the probe.",
								Optional:    true,
							},
							"timeout_seconds": schema.Int64Attribute{
								Description: "Timeout for each probe attempt.",
								Optional:    true,
							},
							"failure_threshold": schema.Int64Attribute{
								Description: "Number of failures before marking as unhealthy.",
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
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

	if dataClient.Functions == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected functions service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Functions
}

// buildCreateFunctionRequest converts the Terraform model to a client request
func (r *Resource) buildCreateFunctionRequest(ctx context.Context, plan ResourceModel) (client.CreateFunctionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := client.CreateFunctionRequest{
		Name:  plan.Name.ValueString(),
		Image: plan.Image.ValueString(),
	}

	// Port
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		req.Port = safeInt32Convert(plan.Port.ValueInt64())
	} else {
		// Use default port 8080 when not specified by user
		req.Port = 8080
	}

	// Environment variables
	if len(plan.Env) > 0 {
		req.Env = make([]client.EnvVar, len(plan.Env))
		for i, env := range plan.Env {
			req.Env[i] = client.EnvVar{
				Name:  env.Name.ValueString(),
				Value: env.Value.ValueString(),
			}
		}
	}

	// Secrets and RegistryAuth are write-only — set from config by the caller.

	// Resources
	if plan.Resources != nil {
		req.Resources = &client.Resources{}
		if !plan.Resources.CPURequest.IsNull() {
			req.Resources.CPURequest = plan.Resources.CPURequest.ValueString()
		}
		if !plan.Resources.CPULimit.IsNull() {
			req.Resources.CPULimit = plan.Resources.CPULimit.ValueString()
		}
		if !plan.Resources.MemoryRequest.IsNull() {
			req.Resources.MemoryRequest = plan.Resources.MemoryRequest.ValueString()
		}
		if !plan.Resources.MemoryLimit.IsNull() {
			req.Resources.MemoryLimit = plan.Resources.MemoryLimit.ValueString()
		}
	}

	// Concurrency
	if plan.Concurrency != nil && !plan.Concurrency.Limit.IsNull() {
		limit := plan.Concurrency.Limit.ValueInt64()
		req.Concurrency = &client.Concurrency{
			Limit: &limit,
		}
	}

	// Health checks
	if plan.HealthChecks != nil {
		req.HealthChecks = &client.HealthChecks{}

		if plan.HealthChecks.Liveness != nil {
			liveness := plan.HealthChecks.Liveness
			req.HealthChecks.Liveness = &client.Probe{
				Path:                liveness.Path.ValueString(),
				Port:                safeInt32Convert(liveness.Port.ValueInt64()),
				InitialDelaySeconds: safeInt32Convert(liveness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       safeInt32Convert(liveness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      safeInt32Convert(liveness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    safeInt32Convert(liveness.FailureThreshold.ValueInt64()),
			}
		}

		if plan.HealthChecks.Readiness != nil {
			readiness := plan.HealthChecks.Readiness
			req.HealthChecks.Readiness = &client.Probe{
				Path:                readiness.Path.ValueString(),
				Port:                safeInt32Convert(readiness.Port.ValueInt64()),
				InitialDelaySeconds: safeInt32Convert(readiness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       safeInt32Convert(readiness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      safeInt32Convert(readiness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    safeInt32Convert(readiness.FailureThreshold.ValueInt64()),
			}
		}
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		tags, tagDiags := expandTagSet(ctx, plan.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return req, diags
		}
		req.Tags = tags
	}

	return req, diags
}

// buildUpdateFunctionRequest converts the Terraform model to an update request
func (r *Resource) buildUpdateFunctionRequest(ctx context.Context, plan ResourceModel) (client.UpdateFunctionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := client.UpdateFunctionRequest{
		Image: plan.Image.ValueString(),
	}

	// Port
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		req.Port = safeInt32Convert(plan.Port.ValueInt64())
	} else {
		// Use default port 8080 when not specified by user
		req.Port = 8080
	}

	// Environment variables
	if len(plan.Env) > 0 {
		req.Env = make([]client.EnvVar, len(plan.Env))
		for i, env := range plan.Env {
			req.Env[i] = client.EnvVar{
				Name:  env.Name.ValueString(),
				Value: env.Value.ValueString(),
			}
		}
	}

	// Secrets and RegistryAuth are write-only — set from config by the caller.

	// Resources
	if plan.Resources != nil {
		req.Resources = &client.Resources{}
		if !plan.Resources.CPURequest.IsNull() {
			req.Resources.CPURequest = plan.Resources.CPURequest.ValueString()
		}
		if !plan.Resources.CPULimit.IsNull() {
			req.Resources.CPULimit = plan.Resources.CPULimit.ValueString()
		}
		if !plan.Resources.MemoryRequest.IsNull() {
			req.Resources.MemoryRequest = plan.Resources.MemoryRequest.ValueString()
		}
		if !plan.Resources.MemoryLimit.IsNull() {
			req.Resources.MemoryLimit = plan.Resources.MemoryLimit.ValueString()
		}
	}

	// Concurrency
	if plan.Concurrency != nil && !plan.Concurrency.Limit.IsNull() {
		limit := plan.Concurrency.Limit.ValueInt64()
		req.Concurrency = &client.Concurrency{
			Limit: &limit,
		}
	}

	// Health checks
	if plan.HealthChecks != nil {
		req.HealthChecks = &client.HealthChecks{}

		if plan.HealthChecks.Liveness != nil {
			liveness := plan.HealthChecks.Liveness
			req.HealthChecks.Liveness = &client.Probe{
				Path:                liveness.Path.ValueString(),
				Port:                safeInt32Convert(liveness.Port.ValueInt64()),
				InitialDelaySeconds: safeInt32Convert(liveness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       safeInt32Convert(liveness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      safeInt32Convert(liveness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    safeInt32Convert(liveness.FailureThreshold.ValueInt64()),
			}
		}

		if plan.HealthChecks.Readiness != nil {
			readiness := plan.HealthChecks.Readiness
			req.HealthChecks.Readiness = &client.Probe{
				Path:                readiness.Path.ValueString(),
				Port:                safeInt32Convert(readiness.Port.ValueInt64()),
				InitialDelaySeconds: safeInt32Convert(readiness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       safeInt32Convert(readiness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      safeInt32Convert(readiness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    safeInt32Convert(readiness.FailureThreshold.ValueInt64()),
			}
		}
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		tags, tagDiags := expandTagSet(ctx, plan.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return req, diags
		}
		req.Tags = tags
	}

	return req, diags
}

// updateModelFromFunction updates the Terraform model with values from the API response
func (r *Resource) updateModelFromFunction(ctx context.Context, model *ResourceModel, function *client.Function) diag.Diagnostics {
	var diags diag.Diagnostics

	// Always set ID (this should always come from API)
	model.ID = types.StringValue(function.Name)

	// TenantID is a computed attribute sourced from the API; always set to a known value.
	model.TenantID = types.StringValue(function.TenantID)

	// Only update Name if API returned a non-empty value, otherwise preserve existing
	if function.Name != "" {
		model.Name = types.StringValue(function.Name)
	}
	// If Name is still unknown/empty, preserve it (don't overwrite planned value)

	// Only update Image if API returned a non-empty value, otherwise preserve existing
	if function.Image != "" {
		model.Image = types.StringValue(function.Image)
	}
	// If Image is still unknown/empty, preserve it (don't overwrite planned value)

	// Status should always be updated from API
	if function.Status != "" {
		model.Status = types.StringValue(function.Status)
	} else {
		model.Status = types.StringValue("Unknown") // Default status
	}

	// Always set all computed attributes to known values - use empty string if not in API response
	if function.URL != "" {
		model.URL = types.StringValue(function.URL)
	} else {
		model.URL = types.StringValue("") // Empty string for missing URL
	}

	if function.LatestReadyRevision != "" {
		model.LatestReadyRevision = types.StringValue(function.LatestReadyRevision)
	} else {
		model.LatestReadyRevision = types.StringValue("") // Empty string for missing revision
	}

	if function.CreatedAt != nil {
		model.CreatedAt = types.StringValue(function.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringValue("") // Empty string for missing created_at
	}

	if function.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(function.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringValue("") // Empty string for missing updated_at
	}

	// Port
	if function.Port != 0 {
		model.Port = types.Int64Value(int64(function.Port))
	} else {
		// If API returns 0 port, use default of 8080
		model.Port = types.Int64Value(8080)
	}

	// Environment variables
	if len(function.Env) > 0 {
		model.Env = make([]EnvVarModel, len(function.Env))
		for i, env := range function.Env {
			model.Env[i] = EnvVarModel{
				Name:  types.StringValue(env.Name),
				Value: types.StringValue(env.Value),
			}
		}
	}

	// Secrets and RegistryAuth are write-only: the API never returns their values, so they
	// are left untouched here and preserved from prior plan/state by the caller.

	// Resources - only update if it was originally specified in the configuration
	// If model.Resources is nil, it means it wasn't specified in config, so keep it nil
	if model.Resources != nil && function.Resources != nil {
		model.Resources = &ResourcesModel{
			CPURequest:    types.StringValue(function.Resources.CPURequest),
			CPULimit:      types.StringValue(function.Resources.CPULimit),
			MemoryRequest: types.StringValue(function.Resources.MemoryRequest),
			MemoryLimit:   types.StringValue(function.Resources.MemoryLimit),
		}
	}

	// Concurrency - only update if it was originally specified in the configuration
	// If model.Concurrency is nil, it means it wasn't specified in config, so keep it nil
	if model.Concurrency != nil && function.Concurrency != nil && function.Concurrency.Limit != nil {
		model.Concurrency = &ConcurrencyModel{
			Limit: types.Int64Value(*function.Concurrency.Limit),
		}
	}

	// Health checks - properly handle health check updates
	if model.HealthChecks != nil {
		if function.HealthChecks != nil {
			// Update liveness probe if it exists
			if model.HealthChecks.Liveness != nil && function.HealthChecks.Liveness != nil {
				liveness := function.HealthChecks.Liveness
				model.HealthChecks.Liveness.Path = types.StringValue(liveness.Path)
				if liveness.Port != 0 {
					model.HealthChecks.Liveness.Port = types.Int64Value(int64(liveness.Port))
				} else {
					model.HealthChecks.Liveness.Port = types.Int64Value(8080)
				}
				model.HealthChecks.Liveness.InitialDelaySeconds = types.Int64Value(int64(liveness.InitialDelaySeconds))
				model.HealthChecks.Liveness.PeriodSeconds = types.Int64Value(int64(liveness.PeriodSeconds))
				model.HealthChecks.Liveness.TimeoutSeconds = types.Int64Value(int64(liveness.TimeoutSeconds))
				model.HealthChecks.Liveness.FailureThreshold = types.Int64Value(int64(liveness.FailureThreshold))
			}
			// Update readiness probe if it exists
			if model.HealthChecks.Readiness != nil && function.HealthChecks.Readiness != nil {
				readiness := function.HealthChecks.Readiness
				model.HealthChecks.Readiness.Path = types.StringValue(readiness.Path)
				if readiness.Port != 0 {
					model.HealthChecks.Readiness.Port = types.Int64Value(int64(readiness.Port))
				} else {
					model.HealthChecks.Readiness.Port = types.Int64Value(8080)
				}
				model.HealthChecks.Readiness.InitialDelaySeconds = types.Int64Value(int64(readiness.InitialDelaySeconds))
				model.HealthChecks.Readiness.PeriodSeconds = types.Int64Value(int64(readiness.PeriodSeconds))
				model.HealthChecks.Readiness.TimeoutSeconds = types.Int64Value(int64(readiness.TimeoutSeconds))
				model.HealthChecks.Readiness.FailureThreshold = types.Int64Value(int64(readiness.FailureThreshold))
			}
		}
	}

	// Tags
	// Preserve the existing planned/state value when the API omits tags entirely.
	if function.Tags != nil {
		tagSet, tagDiags := flattenTags(ctx, function.Tags)
		diags.Append(tagDiags...)
		if !diags.HasError() {
			model.Tags = tagSet
		}
	}

	return diags
}

// Create creates a new function in the ASC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// WriteOnly values (secret values, registry password) are null in plan — read from config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request
	createReq, diags := r.buildCreateFunctionRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	secrets, registryAuth, woDiags := expandWriteOnly(ctx, config)
	resp.Diagnostics.Append(woDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createReq.Secrets = secrets
	createReq.RegistryAuth = registryAuth

	// Create the function
	function, err := r.client.CreateFunction(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating function",
			fmt.Sprintf("Could not create function: %s", err.Error()),
		)
		return
	}

	// Update the model with values from the API response
	resp.Diagnostics.Append(r.updateModelFromFunction(ctx, &plan, function)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	functionName := state.Name.ValueString()

	// Get the function
	function, err := r.client.GetFunction(ctx, functionName)
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			// If function not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting function",
			fmt.Sprintf("Could not get function: %s", err.Error()),
		)
		return
	}

	// Preserve write-only fields — the API never returns them; prior state is the only source.
	priorSecrets := state.Secrets
	priorRegistryAuth := state.RegistryAuth

	// Update the model with values from the API response
	resp.Diagnostics.Append(r.updateModelFromFunction(ctx, &state, function)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Secrets = priorSecrets
	state.RegistryAuth = priorRegistryAuth

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the function in the ASC platform.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config ResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get planned changes
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// WriteOnly values (secret values, registry password) are null in plan — read from config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	functionName := state.Name.ValueString()

	// Use API PUT to update the function in place
	updateReq, diags := r.buildUpdateFunctionRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// PUT replaces the whole function — resend write-only secrets/registry so they survive.
	secrets, registryAuth, woDiags := expandWriteOnly(ctx, config)
	resp.Diagnostics.Append(woDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateReq.Secrets = secrets
	updateReq.RegistryAuth = registryAuth

	function, err := r.client.UpdateFunction(ctx, functionName, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating function",
			fmt.Sprintf("Could not update function '%s': %s", functionName, err.Error()),
		)
		return
	}

	// Update the state with the updated function details
	// Start with the planned state so removals in configuration are preserved,
	// then overlay any values returned by the API.
	updatedState := plan

	resp.Diagnostics.Append(r.updateModelFromFunction(ctx, &updatedState, function)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

// Delete deletes the function in the ASC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	functionName := state.Name.ValueString()

	// Attempt to delete the function
	err := r.client.DeleteFunction(ctx, functionName)
	if err != nil {
		// If the function doesn't exist, that's ok - delete is idempotent
		if errors.Is(err, client.ErrResourceNotFound) {
			// Function doesn't exist, consider delete successful
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting function",
			fmt.Sprintf("Could not delete function '%s': %s", functionName, err.Error()),
		)
		return
	}

	// Delete successful - resource will be automatically removed from state
}

// ImportState imports the state of the function in the ASC platform.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
