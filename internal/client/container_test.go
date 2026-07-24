package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainer_CreateDeployment(t *testing.T) {
	tests := []struct {
		name           string
		container      Container
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful creation",
			container: Container{
				Name: "test-container",
			},
			mockResponse: map[string]any{"data": &Container{
				ID:   "test-id",
				Name: "test-container",
			}},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "conflict error",
			container: Container{
				Name: "existing-container",
			},
			mockResponse:   map[string]any{"error": map[string]any{"code": 409, "message": "container name already exists"}},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "validation error",
			container: Container{
				Name: "",
			},
			mockResponse:   map[string]any{"error": map[string]any{"code": 400, "message": "validation error"}},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "server error",
			container: Container{
				Name: "test-container",
			},
			mockResponse:   map[string]any{"error": map[string]any{"code": 500, "message": "Internal server error"}},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Containers

			container, err := client.CreateDeployment(context.Background(), tt.container)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "test-id", container.ID)
				assert.Equal(t, tt.container.Name, container.Name)
			}
		})
	}
}

func TestContainer_GetDeployment(t *testing.T) {
	tests := []struct {
		name           string
		containerName  string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:          "successful creation",
			containerName: "test-container",
			mockResponse: map[string]any{"data": &Container{
				ID:   "test-id",
				Name: "test-container",
			}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "conflict error",
			containerName:  "nonexistent-container",
			mockResponse:   map[string]any{"error": map[string]any{"code": 404, "message": "not found"}},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			containerName:  "test-container",
			mockResponse:   map[string]any{"error": map[string]any{"code": 500, "message": "Internal server error"}},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Containers

			container, err := client.GetDeployment(context.Background(), tt.containerName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "test-id", container.ID)
				assert.Equal(t, tt.containerName, container.Name)
			}
		})
	}
}

func TestContainer_DeleteDeployment(t *testing.T) {
	tests := []struct {
		name           string
		containerName  string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful creation",
			containerName:  "test-container",
			mockResponse:   map[string]string{"deleted": "test-container"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			containerName:  "nonexistent-vpc",
			mockResponse:   map[string]any{"error": map[string]any{"code": 404, "message": "not found"}},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			containerName:  "test-container",
			mockResponse:   map[string]any{"error": map[string]any{"code": 500, "message": "Internal server error"}},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Containers

			err := client.DeleteDeployment(context.Background(), tt.containerName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
