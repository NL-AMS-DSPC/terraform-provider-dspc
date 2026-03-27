package function

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &FunctionDataSource{}
	_ datasource.DataSourceWithConfigure = &FunctionDataSource{}
)

// FunctionDataSourceClient defines the interface for retrieving function data source information.
type FunctionDataSourceClient interface {
	GetFunction(ctx context.Context, name string) (*client.Function, error)
}

// FunctionDataSource defines the data source implementation.
type FunctionDataSource struct {
	client FunctionDataSourceClient
}

// FunctionDataSourceModel describes the data source data model.
type FunctionDataSourceModel struct {
	Name                types.String        `tfsdk:"name"`
	ID                  types.String        `tfsdk:"id"`
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

// NewFunctionDataSource creates a new FunctionDataSource.
func NewFunctionDataSource() datasource.DataSource {
	return &FunctionDataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *FunctionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *FunctionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific function in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the function to retrieve.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the function.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "The container image for the function.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The port the container listens on.",
				Computed:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL of the function.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the function.",
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
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the environment variable.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The value of the environment variable.",
							Computed:    true,
						},
					},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Description: "Secret environment variables for the function.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the secret.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "The key within the secret to use.",
							Computed:    true,
						},
						"env_name": schema.StringAttribute{
							Description: "The environment variable name to set.",
							Computed:    true,
						},
					},
				},
			},
			"tags": schema.ListNestedAttribute{
				Description: "Tags for the function.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "The tag key.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The tag value.",
							Computed:    true,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"concurrency": schema.ListNestedBlock{
				Description: "Concurrency configuration for the function.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"limit": schema.Int64Attribute{
							Description: "Maximum number of concurrent requests.",
							Computed:    true,
						},
					},
				},
			},
			"resources": schema.ListNestedBlock{
				Description: "Resource limits and requests for the function.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cpu_request": schema.StringAttribute{
							Description: "CPU request.",
							Computed:    true,
						},
						"cpu_limit": schema.StringAttribute{
							Description: "CPU limit.",
							Computed:    true,
						},
						"memory_request": schema.StringAttribute{
							Description: "Memory request.",
							Computed:    true,
						},
						"memory_limit": schema.StringAttribute{
							Description: "Memory limit.",
							Computed:    true,
						},
					},
				},
			},
			"health_checks": schema.ListNestedBlock{
				Description: "Health check configuration for the function.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"liveness": schema.ListNestedBlock{
							Description: "Liveness probe configuration.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"path": schema.StringAttribute{
										Description: "HTTP path for the probe.",
										Computed:    true,
									},
									"port": schema.Int64Attribute{
										Description: "Port for the probe.",
										Computed:    true,
									},
									"initial_delay_seconds": schema.Int64Attribute{
										Description: "Initial delay before probing starts.",
										Computed:    true,
									},
									"period_seconds": schema.Int64Attribute{
										Description: "How often to perform the probe.",
										Computed:    true,
									},
									"timeout_seconds": schema.Int64Attribute{
										Description: "Timeout for each probe attempt.",
										Computed:    true,
									},
									"failure_threshold": schema.Int64Attribute{
										Description: "Number of failures before marking as unhealthy.",
										Computed:    true,
									},
								},
							},
						},
						"readiness": schema.ListNestedBlock{
							Description: "Readiness probe configuration.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"path": schema.StringAttribute{
										Description: "HTTP path for the probe.",
										Computed:    true,
									},
									"port": schema.Int64Attribute{
										Description: "Port for the probe.",
										Computed:    true,
									},
									"initial_delay_seconds": schema.Int64Attribute{
										Description: "Initial delay before probing starts.",
										Computed:    true,
									},
									"period_seconds": schema.Int64Attribute{
										Description: "How often to perform the probe.",
										Computed:    true,
									},
									"timeout_seconds": schema.Int64Attribute{
										Description: "Timeout for each probe attempt.",
										Computed:    true,
									},
									"failure_threshold": schema.Int64Attribute{
										Description: "Number of failures before marking as unhealthy.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *FunctionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Functions == nil {
		resp.Diagnostics.AddError("Unexpected data source configuration error",
			"Expected functions service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Functions
}

// updateModelFromFunction updates the data source model with values from the API response
func (d *FunctionDataSource) updateModelFromFunction(model *FunctionDataSourceModel, function *client.Function) {
	model.ID = types.StringValue(function.Name)

	// For data sources, we can safely set these values directly from API
	if function.Image != "" {
		model.Image = types.StringValue(function.Image)
	} else {
		model.Image = types.StringValue("") // Use empty string instead of null for consistency
	}

	if function.Status != "" {
		model.Status = types.StringValue(function.Status)
	} else {
		model.Status = types.StringValue("Unknown")
	}

	if function.URL != "" {
		model.URL = types.StringValue(function.URL)
	} else {
		model.URL = types.StringValue("")
	}

	if function.LatestReadyRevision != "" {
		model.LatestReadyRevision = types.StringValue(function.LatestReadyRevision)
	} else {
		model.LatestReadyRevision = types.StringValue("")
	}

	if function.CreatedAt != nil {
		model.CreatedAt = types.StringValue(function.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringValue("")
	}

	if function.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(function.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringValue("")
	}

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

	// Resources
	if function.Resources != nil {
		model.Resources = &ResourcesModel{
			CPURequest:    types.StringValue(function.Resources.CPURequest),
			CPULimit:      types.StringValue(function.Resources.CPULimit),
			MemoryRequest: types.StringValue(function.Resources.MemoryRequest),
			MemoryLimit:   types.StringValue(function.Resources.MemoryLimit),
		}
	}

	// Concurrency
	if function.Concurrency != nil && function.Concurrency.Limit != nil {
		model.Concurrency = &ConcurrencyModel{
			Limit: types.Int64Value(*function.Concurrency.Limit),
		}
	}

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

// Read refreshes the Terraform state with the latest data.
func (d *FunctionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FunctionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the function
	function, err := d.client.GetFunction(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting function",
			fmt.Sprintf("Could not get function '%s': %s", config.Name.ValueString(), err.Error()),
		)
		return
	}

	// Update the model with values from the API response
	d.updateModelFromFunction(&config, function)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
