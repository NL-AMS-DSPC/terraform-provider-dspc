// Package provider implements the DSPC Terraform provider for managing resources
// via the DSPC VM Deployer API. It provides resources and data sources for creating,
// reading, and resources, along with an API client for interacting
// with the DSPC service.
package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/resources/blockstorage"
	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/resources/virtualmachine"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure DspcProvider satisfies various provider interfaces.
var _ provider.Provider = &DspcProvider{}

// DspcProvider defines the provider implementation.
type DspcProvider struct {
	version string
}

// DspcProviderModel describes the provider data model.
type DspcProviderModel struct {
	Endpoint  types.String `tfsdk:"endpoint"`
	Timeout   types.Int64  `tfsdk:"timeout"`
	APIKey    types.String `tfsdk:"api_key"`
	Namespace types.String `tfsdk:"namespace"`
}

// Metadata updates the provided metadata with the provider type name and version.
func (p *DspcProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dspc"
	resp.Version = p.version
}

// Schema updates the provider schema with the attributes for the provider.
func (p *DspcProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The DSPC provider manages virtual machines, containers, and storage " +
			"resources across different platforms.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The endpoint URL for the DSPC VM Deployer API. Required - can be set " +
					"via provider config or DSPC_ENDPOINT environment variable.",
				Optional: true,
			},
			"timeout": schema.Int64Attribute{
				Description: "The timeout in seconds for API requests. Defaults to 30.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for authentication with DSPC API. Required - can be set " +
					"via provider config or DSPC_API_KEY environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"namespace": schema.StringAttribute{
				Description: "The name of the namespace where the VM is deployed.",
				Optional:    true,
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for resources and data sources to use.
func (p *DspcProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DspcProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the API client (handles all config extraction and defaults)
	dspcClient, err := newClientFromConfig(config)
	if err != nil {
		resp.Diagnostics.AddError("Provider Configuration Error", err.Error())
		return
	}

	// Store the client in the response data for resources and data sources to use
	resp.ResourceData = dspcClient
	resp.DataSourceData = dspcClient
}

// Resources returns the resources for the provider.
func (p *DspcProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		virtualmachine.NewVMResource,
		blockstorage.NewAttachmentResource,
		blockstorage.NewBlockStorageResource,
	}
}

// DataSources returns the data sources for the provider.
func (p *DspcProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		virtualmachine.NewVMDataSource,
		blockstorage.NewAttachmentDataSource,
		blockstorage.NewDataSource,
	}
}

// New creates a new provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DspcProvider{
			version: version,
		}
	}
}

func newClientFromConfig(config DspcProviderModel) (*client.DspcClient, error) {
	var endpoint, apiKey string
	var timeoutSeconds int64
	var namespace string

	// Extract endpoint with environment fallback
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = os.Getenv("DSPC_ENDPOINT")
	}
	// Validate that endpoint is provided
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required but not provided. Please set the 'endpoint' attribute " +
			"in the provider configuration or set the DSPC_ENDPOINT environment variable")
	}

	if !config.Namespace.IsNull() {
		namespace = config.Namespace.ValueString()
	}
	if namespace == "" {
		namespace = os.Getenv("DSPC_NAMESPACE")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required but not provided. Please set the 'namespace' attribute " +
			"in the provider configuration or set the DSPC_NAMESPACE environment variable")
	}

	// Extract API key with environment fallback
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		apiKey = os.Getenv("DSPC_API_KEY")
	}

	// Validate that API key is provided
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required but not provided. Please set the 'api_key' attribute " +
			"in the provider configuration or set the DSPC_API_KEY environment variable")
	}

	// Extract timeout with defaults
	if !config.Timeout.IsNull() {
		timeoutSeconds = config.Timeout.ValueInt64()
	}
	if timeoutSeconds == 0 {
		if envTimeout := os.Getenv("DSPC_TIMEOUT"); envTimeout != "" {
			if parsedTimeout, err := strconv.ParseInt(envTimeout, 10, 64); err == nil {
				timeoutSeconds = parsedTimeout
			}
		}
		if timeoutSeconds == 0 {
			timeoutSeconds = 30 // default
		}
	}

	return client.NewDspcClient(endpoint, namespace, apiKey, timeoutSeconds), nil
}
