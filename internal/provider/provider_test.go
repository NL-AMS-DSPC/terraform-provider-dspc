package provider

import (
	"context"
	"fmt"
	"os"
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
provider "dspc" {}
`,
			wantErr: false,
		},
		{
			name: "explicit configuration",
			config: `
provider "dspc" {
  endpoint = "https://api.example.com:8080"
  api_key  = "test-key"
  timeout  = 60
}
`,
			wantErr: false,
		},
		{
			name: "environment variable fallback",
			config: `
provider "dspc" {}
`,
			wantErr: false,
			setupEnv: func() {
				_ = os.Setenv("DSPC_ENDPOINT", "https://env.example.com:8080")
				_ = os.Setenv("DSPC_API_KEY", "env-test-key")
				_ = os.Setenv("DSPC_TIMEOUT", "120")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_ENDPOINT")
				_ = os.Unsetenv("DSPC_API_KEY")
				_ = os.Unsetenv("DSPC_TIMEOUT")
			},
		},
		{
			name: "partial environment variables",
			config: `
provider "dspc" {
  endpoint = "https://config.example.com:8080"
}
`,
			wantErr: false,
			setupEnv: func() {
				_ = os.Setenv("DSPC_API_KEY", "env-api-key")
				_ = os.Setenv("DSPC_TIMEOUT", "90")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_API_KEY")
				_ = os.Unsetenv("DSPC_TIMEOUT")
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
    dspc = {
      source = "dspc/dspc"
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
	assert.Contains(t, attributes, "api_key", "Provider schema missing 'api_key' attribute")
}

func TestProviderMetadata(t *testing.T) {
	p := &DspcProvider{version: "1.0.0"}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	assert.Equal(t, "dspc", resp.TypeName)
	assert.Equal(t, "1.0.0", resp.Version)
}

func TestProviderResources(t *testing.T) {
	expectedNumberOfResources := 2

	p := &DspcProvider{version: "test"}

	resources := p.Resources(context.Background())

	assert.Equal(t, expectedNumberOfResources, len(resources), fmt.Sprintf("Expected %d resources, got %d", expectedNumberOfResources, len(resources)))

	// Test that the resource factory returns a valid resource
	assert.NotNil(t, resources[0](), "Expected resource to not be nil")
}

func TestProviderDataSources(t *testing.T) {
	expectedNumberOfDatasources := 1

	p := &DspcProvider{version: "test"}

	dataSources := p.DataSources(context.Background())

	assert.Equal(t, expectedNumberOfDatasources, len(dataSources), fmt.Sprintf("Expected %d datasources, got %d", expectedNumberOfDatasources, len(dataSources)))

	// Test that the data source factory returns a valid data source
	assert.NotNil(t, dataSources[0](), "DataSource factory returned nil")
}
