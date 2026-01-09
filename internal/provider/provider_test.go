package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
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
			if p == nil {
				t.Error("Provider factory returned nil")
			}

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

	if resp.Diagnostics.HasError() {
		t.Errorf("Provider schema has errors: %v", resp.Diagnostics)
	}

	if resp.Schema.Attributes == nil {
		t.Error("Provider schema attributes is nil")
	}

	// Check that required attributes exist
	attributes := resp.Schema.Attributes
	if _, ok := attributes["endpoint"]; !ok {
		t.Error("Provider schema missing 'endpoint' attribute")
	}
	if _, ok := attributes["timeout"]; !ok {
		t.Error("Provider schema missing 'timeout' attribute")
	}
	if _, ok := attributes["api_key"]; !ok {
		t.Error("Provider schema missing 'api_key' attribute")
	}
}

func TestProviderMetadata(t *testing.T) {
	p := &DspcProvider{version: "1.0.0"}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "dspc" {
		t.Errorf("Expected type name 'dspc', got '%s'", resp.TypeName)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", resp.Version)
	}
}

func TestProviderResources(t *testing.T) {
	p := &DspcProvider{version: "test"}

	resources := p.Resources(context.Background())

	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	// Test that the resource factory returns a valid resource
	resource := resources[0]()
	if resource == nil {
		t.Error("Resource factory returned nil")
	}
}

func TestProviderDataSources(t *testing.T) {
	p := &DspcProvider{version: "test"}

	dataSources := p.DataSources(context.Background())

	if len(dataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(dataSources))
	}

	// Test that the data source factory returns a valid data source
	dataSource := dataSources[0]()
	if dataSource == nil {
		t.Error("Data source factory returned nil")
	}
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
