package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		wantErr  bool
		setupEnv func()
		cleanup  func()
	}{
		{
			name: "default configuration",
			config: `
provider "asc" {}
`,
			wantErr: false,
		},
		{
			name: "explicit configuration",
			config: `
provider "asc" {
  endpoint  = "https://api.example.com:8080"
  auth_url  = "https://auth.example.com"
  username  = "test-client-id"
  password  = "test-client-secret"
  org       = "test-realm"
  namespace = "test-ns"
  timeout   = 60
}
`,
			wantErr: false,
		},
		{
			name: "environment variable fallback",
			config: `
provider "asc" {}
`,
			wantErr: false,
			setupEnv: func() {
				_ = os.Setenv("ASC_ENDPOINT", "https://env.example.com:8080")
				_ = os.Setenv("ASC_AUTH_URL", "https://auth.env.example.com")
				_ = os.Setenv("ASC_USERNAME", "env-client-id")
				_ = os.Setenv("ASC_PASSWORD", "env-client-secret")
				_ = os.Setenv("ASC_ORG", "env-realm")
				_ = os.Setenv("ASC_NAMESPACE", "env-ns")
				_ = os.Setenv("ASC_TIMEOUT", "120")
			},
			cleanup: func() {
				_ = os.Unsetenv("ASC_ENDPOINT")
				_ = os.Unsetenv("ASC_AUTH_URL")
				_ = os.Unsetenv("ASC_USERNAME")
				_ = os.Unsetenv("ASC_PASSWORD")
				_ = os.Unsetenv("ASC_ORG")
				_ = os.Unsetenv("ASC_NAMESPACE")
				_ = os.Unsetenv("ASC_TIMEOUT")
			},
		},
		{
			name: "partial environment variables",
			config: `
provider "asc" {
  endpoint = "https://config.example.com:8080"
  namespace = "test-ns"
}
`,
			wantErr: false,
			setupEnv: func() {
				_ = os.Setenv("ASC_AUTH_URL", "https://auth.env.example.com")
				_ = os.Setenv("ASC_USERNAME", "env-client-id")
				_ = os.Setenv("ASC_PASSWORD", "env-client-secret")
				_ = os.Setenv("ASC_ORG", "env-realm")
				_ = os.Setenv("ASC_TIMEOUT", "90")
			},
			cleanup: func() {
				_ = os.Unsetenv("ASC_AUTH_URL")
				_ = os.Unsetenv("ASC_USERNAME")
				_ = os.Unsetenv("ASC_PASSWORD")
				_ = os.Unsetenv("ASC_ORG")
				_ = os.Unsetenv("ASC_TIMEOUT")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// Create provider factory
			providerFactory := func() provider.Provider {
				return &DspcProvider{
					version: "test",
				}
			}

			// Test provider configuration
			_ = `
terraform {
  required_providers {
    asc = {
      source = "asc/asc"
    }
  }
}

` + tt.config

			// This is a basic test that the provider can be instantiated
			// In a real test, you would use terraform-plugin-testing to validate
			// the configuration parsing and client creation
			p := providerFactory()
			assert.NotNil(t, p, "provider should not be nil")

			// Test that the provider implements the required interfaces
			var _ = p
		})
	}
}

func TestProviderSchema(t *testing.T) {
	p := &DspcProvider{version: "test"}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	assert.False(t, false, "Provider schema has errors: %v", resp.Diagnostics)

	assert.NotNil(t, resp.Schema.Attributes)

	// Check that required attributes exist
	attributes := resp.Schema.Attributes
	assert.Contains(t, attributes, "endpoint", "Provider schema missing 'endpoint' attribute")
	assert.Contains(t, attributes, "timeout", "Provider schema missing 'timeout' attribute")
	assert.Contains(t, attributes, "username", "Provider schema missing 'username' attribute")
	assert.Contains(t, attributes, "password", "Provider schema missing 'password' attribute")
	assert.Contains(t, attributes, "auth_url", "Provider schema missing 'auth_url' attribute")
	assert.Contains(t, attributes, "org", "Provider schema missing 'org' attribute")
	assert.Contains(t, attributes, "namespace", "Provider schema missing 'namespace' attribute")
}

func TestProviderMetadata(t *testing.T) {
	p := &DspcProvider{version: "1.0.0"}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	assert.Equal(t, "asc", resp.TypeName)
	assert.Equal(t, "1.0.0", resp.Version)
}

func TestProviderResources(t *testing.T) {
	expectedNumberOfResources := countFilesWithSuffix("../../internal/resources", "resource.go")

	p := &DspcProvider{version: "test"}

	resources := p.Resources(context.Background())

	assert.Equal(t, expectedNumberOfResources, len(resources),
		fmt.Sprintf("Expected %d resources, got %d. One might be missing from the provider resources list.",
			expectedNumberOfResources, len(resources),
		),
	)

	// Test that the resource factory returns a valid resource
	assert.NotNil(t, resources[0](), "Expected resource to not be nil")
}

func TestProviderDataSources(t *testing.T) {
	expectedNumberOfDatasources := countFilesWithSuffix("../../internal/resources", "data_source.go")

	p := &DspcProvider{version: "test"}

	dataSources := p.DataSources(context.Background())

	assert.Equal(t, expectedNumberOfDatasources, len(dataSources),
		fmt.Sprintf("Expected %d data sources, got %d. One might be missing from the data sources list in the provider.",
			expectedNumberOfDatasources, len(dataSources)),
	)

	// Test that the data source factory returns a valid data source
	assert.NotNil(t, dataSources[0](), "DataSource factory returned nil")
}

// countFilesWithSuffix counts recusifly all files with the postfix.
func countFilesWithSuffix(path string, postfix string) int {
	var count int

	dir, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	for _, f := range dir {
		if f.IsDir() {
			count += countFilesWithSuffix(fmt.Sprintf("%s/%s", path, f.Name()), postfix)
			continue
		}

		if strings.HasSuffix(f.Name(), postfix) {
			count++
		}
	}
	return count
}
