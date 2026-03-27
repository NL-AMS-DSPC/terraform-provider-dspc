package function

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &FunctionResource{}
	_ resource.ResourceWithConfigure   = &FunctionResource{}
	_ resource.ResourceWithImportState = &FunctionResource{}
)

// FunctionResourceClient defines the interface for managing function resources.
// It provides methods to create, delete, retrieve, and list functions.
type FunctionResourceClient interface {
	CreateFunction(ctx context.Context, req client.CreateFunctionRequest) (*client.Function, error)
	UpdateFunction(ctx context.Context, name string, req client.UpdateFunctionRequest) (*client.Function, error)
	DeleteFunction(ctx context.Context, name string) error
	GetFunction(ctx context.Context, name string) (*client.Function, error)
	ListFunctions(ctx context.Context) ([]*client.Function, error)
}

// FunctionResource defines the resource implementation.
type FunctionResource struct {
	client FunctionResourceClient
}

// EnvVarModel represents an environment variable
type EnvVarModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// SecretEnvVarModel represents a secret environment variable
type SecretEnvVarModel struct {
	Name    types.String `tfsdk:"name"`
	Key     types.String `tfsdk:"key"`
	EnvName types.String `tfsdk:"env_name"`
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

// FunctionResourceModel describes the resource data model.
type FunctionResourceModel struct {
	ID                  types.String        `tfsdk:"id"`
	Name                types.String        `tfsdk:"name"`
	Image               types.String        `tfsdk:"image"`
	Port                types.Int64         `tfsdk:"port"`
	Env                 []EnvVarModel       `tfsdk:"env"`
	Secrets             []SecretEnvVarModel `tfsdk:"secrets"`
	Resources           *ResourcesModel     `tfsdk:"resources"`
	Concurrency         *ConcurrencyModel   `tfsdk:"concurrency"`
	HealthChecks        *HealthChecksModel  `tfsdk:"health_checks"`
	Tags                []TagModel          `tfsdk:"tags"`
	URL                 types.String        `tfsdk:"url"`
	Status              types.String        `tfsdk:"status"`
	LatestReadyRevision types.String        `tfsdk:"latest_ready_revision"`
	CreatedAt           types.String        `tfsdk:"created_at"`
	UpdatedAt           types.String        `tfsdk:"updated_at"`
}

// NewFunctionResource creates a new FunctionResource.
func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *FunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *FunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a function in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the function.",
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
				Description: "Secret environment variables for the function.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the secret.",
							Required:    true,
						},
						"key": schema.StringAttribute{
							Description: "The key within the secret to use.",
							Required:    true,
						},
						"env_name": schema.StringAttribute{
							Description: "The environment variable name to set.",
							Required:    true,
						},
					},
				},
			},
			"tags": schema.ListNestedAttribute{
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
func (r *FunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *FunctionResource) buildCreateFunctionRequest(plan FunctionResourceModel) client.CreateFunctionRequest {
	req := client.CreateFunctionRequest{
		Name:  plan.Name.ValueString(),
		Image: plan.Image.ValueString(),
	}

	// Port
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		req.Port = int32(plan.Port.ValueInt64())
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

	// Secrets
	if len(plan.Secrets) > 0 {
		req.Secrets = make([]client.SecretEnvVar, len(plan.Secrets))
		for i, secret := range plan.Secrets {
			req.Secrets[i] = client.SecretEnvVar{
				Name:    secret.Name.ValueString(),
				Key:     secret.Key.ValueString(),
				EnvName: secret.EnvName.ValueString(),
			}
		}
	}

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
				Port:                int32(liveness.Port.ValueInt64()),
				InitialDelaySeconds: int32(liveness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       int32(liveness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      int32(liveness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    int32(liveness.FailureThreshold.ValueInt64()),
			}
		}

		if plan.HealthChecks.Readiness != nil {
			readiness := plan.HealthChecks.Readiness
			req.HealthChecks.Readiness = &client.Probe{
				Path:                readiness.Path.ValueString(),
				Port:                int32(readiness.Port.ValueInt64()),
				InitialDelaySeconds: int32(readiness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       int32(readiness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      int32(readiness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    int32(readiness.FailureThreshold.ValueInt64()),
			}
		}
	}

	// Tags
	if len(plan.Tags) > 0 {
		req.Tags = make([]client.Tag, len(plan.Tags))
		for i, tag := range plan.Tags {
			req.Tags[i] = client.Tag{
				Key:   tag.Key.ValueString(),
				Value: tag.Value.ValueString(),
			}
		}
	}

	return req
}

// buildUpdateFunctionRequest converts the Terraform model to an update request
func (r *FunctionResource) buildUpdateFunctionRequest(plan FunctionResourceModel) client.UpdateFunctionRequest {
	req := client.UpdateFunctionRequest{
		Image: plan.Image.ValueString(),
	}

	// Port
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		req.Port = int32(plan.Port.ValueInt64())
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

	// Secrets
	if len(plan.Secrets) > 0 {
		req.Secrets = make([]client.SecretEnvVar, len(plan.Secrets))
		for i, secret := range plan.Secrets {
			req.Secrets[i] = client.SecretEnvVar{
				Name:    secret.Name.ValueString(),
				Key:     secret.Key.ValueString(),
				EnvName: secret.EnvName.ValueString(),
			}
		}
	}

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
				Port:                int32(liveness.Port.ValueInt64()),
				InitialDelaySeconds: int32(liveness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       int32(liveness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      int32(liveness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    int32(liveness.FailureThreshold.ValueInt64()),
			}
		}

		if plan.HealthChecks.Readiness != nil {
			readiness := plan.HealthChecks.Readiness
			req.HealthChecks.Readiness = &client.Probe{
				Path:                readiness.Path.ValueString(),
				Port:                int32(readiness.Port.ValueInt64()),
				InitialDelaySeconds: int32(readiness.InitialDelaySeconds.ValueInt64()),
				PeriodSeconds:       int32(readiness.PeriodSeconds.ValueInt64()),
				TimeoutSeconds:      int32(readiness.TimeoutSeconds.ValueInt64()),
				FailureThreshold:    int32(readiness.FailureThreshold.ValueInt64()),
			}
		}
	}

	// Tags
	if len(plan.Tags) > 0 {
		req.Tags = make([]client.Tag, len(plan.Tags))
		for i, tag := range plan.Tags {
			req.Tags[i] = client.Tag{
				Key:   tag.Key.ValueString(),
				Value: tag.Value.ValueString(),
			}
		}
	}

	return req
}

// updateModelFromFunction updates the Terraform model with values from the API response
func (r *FunctionResource) updateModelFromFunction(model *FunctionResourceModel, function *client.Function) {
	// Always set ID (this should always come from API)
	model.ID = types.StringValue(function.Name)

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
		model.Port = types.Int64Null()
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

	// Secrets
	if len(function.Secrets) > 0 {
		model.Secrets = make([]SecretEnvVarModel, len(function.Secrets))
		for i, secret := range function.Secrets {
			model.Secrets[i] = SecretEnvVarModel{
				Name:    types.StringValue(secret.Name),
				Key:     types.StringValue(secret.Key),
				EnvName: types.StringValue(secret.EnvName),
			}
		}
	}

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

	// Health checks - this is complex and would need proper handling
	// For now, we'll leave it as is since the API response should preserve what was sent

	// Tags
	if len(function.Tags) > 0 {
		model.Tags = make([]TagModel, len(function.Tags))
		for i, tag := range function.Tags {
			model.Tags[i] = TagModel{
				Key:   types.StringValue(tag.Key),
				Value: types.StringValue(tag.Value),
			}
		}
	}
}

// Create creates a new function in the DSPC platform.
func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FunctionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the create request
	createReq := r.buildCreateFunctionRequest(plan)

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
	r.updateModelFromFunction(&plan, function)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FunctionResourceModel

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

	// Update the model with values from the API response
	r.updateModelFromFunction(&state, function)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the function in the DSPC platform.
func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FunctionResourceModel

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

	functionName := state.Name.ValueString()

	// Use API PUT to update the function in place
	updateReq := r.buildUpdateFunctionRequest(plan)
	function, err := r.client.UpdateFunction(ctx, functionName, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating function",
			fmt.Sprintf("Could not update function '%s': %s", functionName, err.Error()),
		)
		return
	}

	// Update the state with the updated function details
	r.updateModelFromFunction(&plan, function)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the function in the DSPC platform.
func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FunctionResourceModel

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

// ImportState imports the state of the function in the DSPC platform.
func (r *FunctionResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
