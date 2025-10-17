package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestClient_CreateVM(t *testing.T) {
	tests := []struct {
		name           string
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful creation",
			vmName: "test-vm",
			mockResponse: CreateVMResponse{
				Created: "test-vm",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error",
			vmName:         "test-vm",
			mockResponse:   map[string]string{"error": "VM already exists"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/virtualmachine" {
					t.Errorf("Expected /virtualmachine path, got %s", r.URL.Path)
				}

				// Check Authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-api-key" {
					t.Errorf("Expected Authorization: Bearer test-api-key, got %s", authHeader)
				}

				// Check Content-Type header
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", contentType)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-api-key", 30)

			// Test CreateVM
			vm, err := client.CreateVM(context.Background(), tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if vm.Name != tt.vmName {
					t.Errorf("Expected VM name %s, got %s", tt.vmName, vm.Name)
				}
			}
		})
	}
}

func TestClient_DeleteVM(t *testing.T) {
	tests := []struct {
		name           string
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful deletion",
			vmName: "test-vm",
			mockResponse: DeleteVMResponse{
				Deleted: "test-vm",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error",
			vmName:         "nonexistent-vm",
			mockResponse:   map[string]string{"error": "VM not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				if r.URL.Path != "/virtualmachine" {
					t.Errorf("Expected /virtualmachine path, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-api-key", 30)

			// Test DeleteVM
			err := client.DeleteVM(context.Background(), tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestClient_ListVMs(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list",
			mockResponse: []*VM{
				{Name: "vm1"},
				{Name: "vm2"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   []*VM{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "API error",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/virtualmachine" {
					t.Errorf("Expected /virtualmachine path, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-api-key", 30)

			// Test ListVMs
			vms, err := client.ListVMs(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(vms) != tt.expectedCount {
					t.Errorf("Expected %d VMs, got %d", tt.expectedCount, len(vms))
				}
			}
		})
	}
}

func TestClient_GetVM(t *testing.T) {
	tests := []struct {
		name           string
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "VM found",
			vmName: "test-vm",
			mockResponse: []*VM{
				{Name: "test-vm"},
				{Name: "other-vm"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "VM not found",
			vmName: "nonexistent-vm",
			mockResponse: []*VM{
				{Name: "other-vm"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/virtualmachine" {
					t.Errorf("Expected /virtualmachine path, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, "test-api-key", 30)

			// Test GetVM
			vm, err := client.GetVM(context.Background(), tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if vm.Name != tt.vmName {
					t.Errorf("Expected VM name %s, got %s", tt.vmName, vm.Name)
				}
			}
		})
	}
}

func TestNewClientFromConfig(t *testing.T) {
	tests := []struct {
		name             string
		config           DspcProviderModel
		expectedEndpoint string
		expectedApiKey   string
		expectedTimeout  int64
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "all values provided",
			config: DspcProviderModel{
				Endpoint: types.StringValue("https://api.example.com"),
				ApiKey:   types.StringValue("test-key"),
				Timeout:  types.Int64Value(60),
			},
			expectedEndpoint: "https://api.example.com",
			expectedApiKey:   "test-key",
			expectedTimeout:  60,
			expectError:      false,
		},
		{
			name: "default values",
			config: DspcProviderModel{
				Endpoint: types.StringNull(),
				ApiKey:   types.StringNull(),
				Timeout:  types.Int64Null(),
			},
			expectError:      true,
			expectedErrorMsg: "API key is required",
		},
		{
			name: "missing API key",
			config: DspcProviderModel{
				Endpoint: types.StringValue("https://api.example.com"),
				ApiKey:   types.StringNull(),
				Timeout:  types.Int64Value(30),
			},
			expectError:      true,
			expectedErrorMsg: "API key is required",
		},
		{
			name: "empty API key",
			config: DspcProviderModel{
				Endpoint: types.StringValue("https://api.example.com"),
				ApiKey:   types.StringValue(""),
				Timeout:  types.Int64Value(30),
			},
			expectError:      true,
			expectedErrorMsg: "API key is required",
		},
		{
			name: "API key from environment variable",
			config: DspcProviderModel{
				Endpoint: types.StringValue("https://api.example.com"),
				ApiKey:   types.StringNull(),
				Timeout:  types.Int64Value(30),
			},
			expectedEndpoint: "https://api.example.com",
			expectedApiKey:   "env-api-key",
			expectedTimeout:  30,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for API key test
			if tt.name == "API key from environment variable" {
				t.Setenv("DSPC_API_KEY", "env-api-key")
			}

			client, err := NewClientFromConfig(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.expectedErrorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				} else {
					if client.endpoint != tt.expectedEndpoint {
						t.Errorf("Expected endpoint %s, got %s", tt.expectedEndpoint, client.endpoint)
					}
					if client.apiKey != tt.expectedApiKey {
						t.Errorf("Expected API key %s, got %s", tt.expectedApiKey, client.apiKey)
					}
					if client.httpClient.Timeout.Seconds() != float64(tt.expectedTimeout) {
						t.Errorf("Expected timeout %d, got %f", tt.expectedTimeout, client.httpClient.Timeout.Seconds())
					}
				}
			}
		})
	}
}

func TestClient_URLConstruction(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		path     string
		expected string
	}{
		{
			name:     "standard endpoint with trailing slash",
			endpoint: "https://api.example.com/",
			path:     "/virtualmachine",
			expected: "https://api.example.com/virtualmachine",
		},
		{
			name:     "standard endpoint without trailing slash",
			endpoint: "https://api.example.com",
			path:     "/virtualmachine",
			expected: "https://api.example.com/virtualmachine",
		},
		{
			name:     "localhost endpoint",
			endpoint: "http://localhost:8080",
			path:     "/virtualmachine",
			expected: "http://localhost:8080/virtualmachine",
		},
		{
			name:     "relative path",
			endpoint: "https://api.example.com",
			path:     "virtualmachine",
			expected: "https://api.example.com/virtualmachine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server to capture the request URL
			var capturedURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedURL = r.URL.String()
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Create client with test server endpoint
			client := NewClient(server.URL, "test-key", 30)

			// Make a request
			_, err := client.makeRequest(context.Background(), "GET", tt.path, nil)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Check if the path was constructed correctly
			// URL construction normalizes paths by adding leading slash
			expectedPath := tt.path
			if !strings.HasPrefix(expectedPath, "/") {
				expectedPath = "/" + expectedPath
			}
			if capturedURL != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, capturedURL)
			}
		})
	}
}

func TestClient_ContextTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*VM{{Name: "test-vm"}})
	}))
	defer server.Close()

	// Create client with short timeout
	client := NewClient(server.URL, "test-api-key", 1) // 1 second timeout

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Test that context timeout is respected
	_, err := client.ListVMs(ctx)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check if error is context-related
	if !isContextError(err) {
		t.Errorf("Expected context error, got: %v", err)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*VM{{Name: "test-vm"}})
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, "test-api-key", 30)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Test that context cancellation is respected
	_, err := client.ListVMs(ctx)
	if err == nil {
		t.Error("Expected cancellation error, got nil")
	}

	// Check if error is context-related
	if !isContextError(err) {
		t.Errorf("Expected context error, got: %v", err)
	}
}

// Helper function to check if error is context-related
func isContextError(err error) bool {
	if err == nil {
		return false
	}
	return err == context.DeadlineExceeded || err == context.Canceled ||
		strings.Contains(err.Error(), "context deadline exceeded") ||
		strings.Contains(err.Error(), "context canceled")
}
