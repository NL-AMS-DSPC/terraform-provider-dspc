package provider

import (
	"context"

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
	Endpoint types.String `tfsdk:"endpoint"`
	Timeout  types.Int64  `tfsdk:"timeout"`
	ApiKey   types.String `tfsdk:"api_key"`
}

func (p *DspcProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dspc"
	resp.Version = p.version
}

func (p *DspcProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The DSPC provider manages virtual machines, containers, and storage resources across different platforms.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The endpoint URL for the DSPC VM Deployer API. Defaults to 'http://localhost:8080'.",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "The timeout in seconds for API requests. Defaults to 30.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for authentication with DSPC API.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DspcProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DspcProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get configuration values (defaults handled in NewClientFromConfig)
	endpoint := ""
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	timeout := int64(0)
	if !config.Timeout.IsNull() {
		timeout = config.Timeout.ValueInt64()
	}

	apiKey := ""
	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	// Create the API client
	client := NewClientFromConfig(endpoint, apiKey, timeout)

	// Store the client in the response data for resources and data sources to use
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *DspcProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVirtualMachineResource,
	}
}

func (p *DspcProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVirtualMachineDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DspcProvider{
			version: version,
		}
	}
}
