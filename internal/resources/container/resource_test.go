package container

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

const (
	containerPath = "/api/containers/v1/namespaces/test-ns/deployments"
)

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		container      client.Container
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful creation",
			container: client.Container{
				Name: "test-container",
			},
			mockResponse: map[string]any{"data": &client.Container{
				Name: "test-container",
			}},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "API error - conflict",
			container: client.Container{
				Name: "existing-container",
			},
			mockResponse:   map[string]string{"error": "Container name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer authServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			containerResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}

			container, err := containerResource.client.CreateDeployment(
				context.Background(),
				tt.container,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.container.Name, container.Name)
			}
		})
	}
}

func TestResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		containerName  string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			containerName:  "test-container",
			mockResponse:   map[string]string{"deleted": "test-container"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error - not found",
			containerName:  "nonexistent-container",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer authServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			containerResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}
			err := containerResource.client.DeleteDeployment(context.Background(), tt.containerName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResource_ImportState(t *testing.T) {
	tests := []struct {
		name           string
		importID       string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "successful import",
			importID: "test-container",
			mockResponse: map[string]any{"data": &client.Container{
				Name: "test-container",
			}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "import non-existent container",
			importID:       "nonexistent-container",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "API error during import",
			importID:       "test-container",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer authServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != containerPath+"/"+tt.importID {
					t.Fatalf("Expected %s path, got %s", containerPath+"/"+tt.importID, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			containerResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}
			container, err := containerResource.client.GetDeployment(context.Background(), tt.importID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.importID, container.Name)
			}
		})
	}
}
