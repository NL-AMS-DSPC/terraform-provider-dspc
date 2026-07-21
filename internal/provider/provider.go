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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/blockstorage"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/cluster"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/container"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/filestorage"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/function"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/group"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/managed_database/mssql"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/managed_database/postgresql"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/objectstorage"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/role"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/securitygroup"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/securitygroupattachment"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/securityrule"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/subnet"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/virtualmachine"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/vmgroup"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/vpc"
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
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	AuthURL   types.String `tfsdk:"auth_url"`
	Org       types.String `tfsdk:"org"`
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
			"username": schema.StringAttribute{
				Description: "Username for authentication. Required - can be set " +
					"via provider config or DSPC_USERNAME environment variable.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "Password for authentication. Required - can be set " +
					"via provider config or DSPC_PASSWORD environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"auth_url": schema.StringAttribute{
				Description: "Authentication service URL. Required - can be set " +
					"via provider config or DSPC_AUTH_URL environment variable.",
				Optional: true,
			},
			"org": schema.StringAttribute{
				Description: "Organization for authentication. Required - can be set " +
					"via provider config or DSPC_ORG environment variable.",
				Optional: true,
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
		vmgroup.NewVMGroupResource,
		function.NewFunctionResource,
		blockstorage.NewAttachmentResource,
		blockstorage.NewBlockStorageResource,
		container.NewResource,
		cluster.NewResource,
		vpc.NewResource,
		subnet.NewResource,
		role.NewResource,
		group.NewResource,
		group.NewMemberResource,
		group.NewRoleResource,
		securitygroup.NewResource,
		securityrule.NewResource,
		securitygroupattachment.NewResource,
		mssql.NewResource,
		postgresql.NewResource,
		filestorage.NewResource,
		filestorage.NewAccessResource,
		objectstorage.NewResource,
	}
}

// DataSources returns the data sources for the provider.
func (p *DspcProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		virtualmachine.NewVMDataSource,
		vmgroup.NewVMGroupDataSource,
		function.NewFunctionDataSource,
		blockstorage.NewAttachmentDataSource,
		blockstorage.NewDataSource,
		container.NewDataSource,
		vpc.NewDataSource,
		subnet.NewDataSource,
		role.NewDataSource,
		group.NewDataSource,
		securitygroup.NewDataSource,
		securityrule.NewDataSource,
		securitygroupattachment.NewDataSource,
		mssql.NewDataSource,
		postgresql.NewDataSource,
		filestorage.NewDataSource,
		filestorage.NewAccessDataSource,
		objectstorage.NewDataSource,
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
	var endpoint, username, password, authURL, org, namespace string
	var timeoutSeconds int64

	// Extract endpoint with environment fallback
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = os.Getenv("DSPC_ENDPOINT")
	}
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

	// Extract username (client_id) with environment fallback
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if username == "" {
		username = os.Getenv("DSPC_USERNAME")
	}
	if username == "" {
		return nil, fmt.Errorf("username is required but not provided. Please set the 'username' attribute " +
			"in the provider configuration or set the DSPC_USERNAME environment variable")
	}

	// Extract password (client_secret) with environment fallback
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	if password == "" {
		password = os.Getenv("DSPC_PASSWORD")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required but not provided. Please set the 'password' attribute " +
			"in the provider configuration or set the DSPC_PASSWORD environment variable")
	}

	// Extract auth_url with environment fallback
	if !config.AuthURL.IsNull() {
		authURL = config.AuthURL.ValueString()
	}
	if authURL == "" {
		authURL = os.Getenv("DSPC_AUTH_URL")
	}
	if authURL == "" {
		return nil, fmt.Errorf("auth_url is required but not provided. Please set the 'auth_url' attribute " +
			"in the provider configuration or set the DSPC_AUTH_URL environment variable")
	}

	// Extract org (realm) with environment fallback
	if !config.Org.IsNull() {
		org = config.Org.ValueString()
	}
	if org == "" {
		org = os.Getenv("DSPC_ORG")
	}
	if org == "" {
		return nil, fmt.Errorf("org is required but not provided. Please set the 'org' attribute " +
			"in the provider configuration or set the DSPC_ORG environment variable")
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

	return client.NewDspcClient(endpoint, namespace, username, password, authURL, org, timeoutSeconds), nil
}
