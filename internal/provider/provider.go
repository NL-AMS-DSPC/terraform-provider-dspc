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
		},
	}
}

func (p *DspcProvider) Configure(_ context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Empty implementation for docs generation
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