package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

			client := newTestDspcClient(server.URL, authServer.URL).Containers

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

			client := newTestDspcClient(server.URL, authServer.URL).Containers

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

func TestContainer_PatchDeployment(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name           string
		containerName  string
		req            PatchTagsRequest
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:          "successful patch",
			containerName: "test-container",
			req:           PatchTagsRequest{Tags: []PatchTagDTO{{Key: "env", Value: strPtr("prod")}}},
			mockResponse: map[string]any{"data": &Container{
				Name: "test-container",
				Tags: []ContainerTag{{Key: "env", Value: "prod"}},
			}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			containerName:  "nonexistent-container",
			req:            PatchTagsRequest{Tags: []PatchTagDTO{{Key: "env", Value: strPtr("prod")}}},
			mockResponse:   map[string]any{"error": map[string]any{"code": 404, "message": "not found"}},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			containerName:  "test-container",
			req:            PatchTagsRequest{Tags: []PatchTagDTO{{Key: "env", Value: strPtr("prod")}}},
			mockResponse:   map[string]any{"error": map[string]any{"code": 500, "message": "Internal server error"}},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Containers

			container, err := client.PatchDeployment(context.Background(), tt.containerName, tt.req)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.containerName, container.Name)
				assert.Equal(t, "prod", container.Tags[0].Value)
			}
		})
	}
}

// TestContainer_PatchDeployment_RequestBody asserts the HTTP method, path, and the
// exact tag-mutation wire format: an upsert carries its string value and a deletion
// marshals as "value":null.
func TestContainer_PatchDeployment_RequestBody(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	authServer := createMockAuthServer()
	defer authServer.Close()

	var (
		gotMethod string
		gotPath   string
		gotBody   string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"name":"test-container"}}`))
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).Containers

	req := PatchTagsRequest{Tags: []PatchTagDTO{
		{Key: "env", Value: strPtr("prod")},
		{Key: "old", Value: nil},
	}}
	_, err := client.PatchDeployment(context.Background(), "test-container", req)
	assert.NoError(t, err)

	assert.Equal(t, http.MethodPatch, gotMethod)
	assert.Equal(t, "/api/containers/v1/deployments/test-container", gotPath)
	assert.True(t, strings.Contains(gotBody, `"key":"env","value":"prod"`), "upsert should carry its string value, got: %s", gotBody)
	assert.True(t, strings.Contains(gotBody, `"key":"old","value":null`), "deletion should marshal as value:null, got: %s", gotBody)
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

			client := newTestDspcClient(server.URL, authServer.URL).Containers

			err := client.DeleteDeployment(context.Background(), tt.containerName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
