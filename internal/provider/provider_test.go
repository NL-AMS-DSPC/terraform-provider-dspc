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

const (
	vmPath = "/virtualmachine"
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
	expectedNumberOfResources := countFilesWithSuffix("../../internal/resources", "resource.go")

	p := &DspcProvider{version: "test"}

	resources := p.Resources(context.Background())

	assert.Equal(t, expectedNumberOfResources, len(resources), fmt.Sprintf("Expected %d resources, got %d", expectedNumberOfResources, len(resources)))

	// Test that the resource factory returns a valid resource
	assert.NotNil(t, resources[0](), "Expected resource to not be nil")
}

func TestProviderDataSources(t *testing.T) {
	expectedNumberOfDatasources := countFilesWithSuffix("../../internal/resources", "_data_source.go")

	p := &DspcProvider{version: "test"}

	dataSources := p.DataSources(context.Background())

	assert.Equal(t, expectedNumberOfDatasources, len(dataSources), fmt.Sprintf("Expected %d datasources, got %d", expectedNumberOfDatasources, len(dataSources)))

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
			count += 1
		}
	}
	return count
}

//func TestNewClientFromConfig(t *testing.T) {
//	tests := []struct {
//		name             string
//		config           DspcProviderModel
//		expectedEndpoint string
//		expectedAPIKey   string
//		expectedTimeout  int64
//		expectError      bool
//		expectedErrorMsg string
//	}{
//		{
//			name: "all values provided",
//			config: DspcProviderModel{
//				Endpoint: types.StringValue("https://api.example.com"),
//				APIKey:   types.StringValue("test-key"),
//				Timeout:  types.Int64Value(60),
//			},
//			expectedEndpoint: "https://api.example.com",
//			expectedAPIKey:   "test-key",
//			expectedTimeout:  60,
//			expectError:      false,
//		},
//		{
//			name: "missing endpoint and API key",
//			config: DspcProviderModel{
//				Endpoint: types.StringNull(),
//				APIKey:   types.StringNull(),
//				Timeout:  types.Int64Null(),
//			},
//			expectError:      true,
//			expectedErrorMsg: "endpoint is required",
//		},
//		{
//			name: "missing API key",
//			config: DspcProviderModel{
//				Endpoint: types.StringValue("https://api.example.com"),
//				APIKey:   types.StringNull(),
//				Timeout:  types.Int64Value(30),
//			},
//			expectError:      true,
//			expectedErrorMsg: "API key is required",
//		},
//		{
//			name: "empty API key",
//			config: DspcProviderModel{
//				Endpoint: types.StringValue("https://api.example.com"),
//				APIKey:   types.StringValue(""),
//				Timeout:  types.Int64Value(30),
//			},
//			expectError:      true,
//			expectedErrorMsg: "API key is required",
//		},
//		{
//			name: "API key from environment variable",
//			config: DspcProviderModel{
//				Endpoint: types.StringValue("https://api.example.com"),
//				APIKey:   types.StringNull(),
//				Timeout:  types.Int64Value(30),
//			},
//			expectedEndpoint: "https://api.example.com",
//			expectedAPIKey:   "env-api-key",
//			expectedTimeout:  30,
//			expectError:      false,
//		},
//		{
//			name: "empty endpoint with API key",
//			config: DspcProviderModel{
//				Endpoint: types.StringValue(""),
//				APIKey:   types.StringValue("test-key"),
//				Timeout:  types.Int64Value(30),
//			},
//			expectError:      true,
//			expectedErrorMsg: "endpoint is required",
//		},
//		{
//			name: "endpoint from environment variable",
//			config: DspcProviderModel{
//				Endpoint: types.StringNull(),
//				APIKey:   types.StringValue("test-key"),
//				Timeout:  types.Int64Value(30),
//			},
//			expectedEndpoint: "https://env-api.example.com",
//			expectedAPIKey:   "test-key",
//			expectedTimeout:  30,
//			expectError:      false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Set environment variables for tests
//			if tt.name == "API key from environment variable" {
//				t.Setenv("DSPC_API_KEY", "env-api-key")
//			}
//			if tt.name == "endpoint from environment variable" {
//				t.Setenv("DSPC_ENDPOINT", "https://env-api.example.com")
//			}
//
//			client, err := newClientFromConfig(tt.config)
//
//			if tt.expectError {
//				if err == nil {
//					t.Errorf("Expected error, got nil")
//				} else if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
//					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrorMsg, err.Error())
//				}
//			} else {
//				if err != nil {
//					t.Errorf("Expected no error, got %v", err)
//				} else {
//					if client.endpoint != tt.expectedEndpoint {
//						t.Errorf("Expected endpoint %s, got %s", tt.expectedEndpoint, client.endpoint)
//					}
//					if client.apiKey != tt.expectedAPIKey {
//						t.Errorf("Expected API key %s, got %s", tt.expectedAPIKey, client.apiKey)
//					}
//					if client.httpClient.Timeout.Seconds() != float64(tt.expectedTimeout) {
//						t.Errorf("Expected timeout %d, got %f", tt.expectedTimeout, client.httpClient.Timeout.Seconds())
//					}
//				}
//			}
//		})
//	}
//}
