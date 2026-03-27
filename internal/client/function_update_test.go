package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionClient_UpdateFunction(t *testing.T) {
	// Create a test server for both auth and API requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth requests (POST to /auth/realms/*)
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/auth/realms/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token":"test-token","expires_in":3600,"token_type":"Bearer"}`))
			return
		}

		// Handle function update requests (PUT to functions endpoint)
		if r.Method == http.MethodPut && r.URL.Path == "/api/v1/v1/functions/test-function" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"updated":"test-function"}`))
			return
		}

		// Handle function get requests (GET to functions endpoint)
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/v1/functions/test-function" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"name": "test-function",
				"image": "updated-image:latest",
				"port": 8080,
				"status": "Running",
				"env": [{"name":"ENV_VAR","value":"updated"}],
				"tags": [{"key":"environment","value":"test"}]
			}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create auth manager with test server URL
	authManager := newAuthManager(&http.Client{}, server.URL+"/auth/realms/test", "test-org", "test-user", "test-pass")

	// Create function client
	functionClient := newFunctionClient(server.URL, "test-ns", "/api/v1", authManager, &http.Client{})

	// Create update request
	updateReq := UpdateFunctionRequest{
		Image: "updated-image:latest",
		Port:  8080,
		Env: []EnvVar{
			{Name: "ENV_VAR", Value: "updated"},
		},
		Tags: []Tag{
			{Key: "environment", Value: "test"},
		},
	}

	// Test that update function succeeds
	ctx := context.Background()
	function, err := functionClient.UpdateFunction(ctx, "test-function", updateReq)

	// Should not return an error
	require.NoError(t, err, "UpdateFunction should succeed")
	require.NotNil(t, function, "Function should not be nil")

	// Verify the returned function has expected values
	assert.Equal(t, "test-function", function.Name)
	assert.Equal(t, "updated-image:latest", function.Image)
	assert.Equal(t, int32(8080), function.Port)
	assert.Equal(t, "Running", function.Status)
	assert.Len(t, function.Env, 1)
	assert.Equal(t, "ENV_VAR", function.Env[0].Name)
	assert.Equal(t, "updated", function.Env[0].Value)
	assert.Len(t, function.Tags, 1)
	assert.Equal(t, "environment", function.Tags[0].Key)
	assert.Equal(t, "test", function.Tags[0].Value)
}

func TestFunctionClient_UpdateFunction_NotFound(t *testing.T) {
	// Create a test server that returns 404 for update requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth requests (POST to /auth/realms/*)
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/auth/realms/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token":"test-token","expires_in":3600,"token_type":"Bearer"}`))
			return
		}

		// Return 404 for all other requests
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Function not found"))
	}))
	defer server.Close()

	// Create auth manager with test server URL
	authManager := newAuthManager(&http.Client{}, server.URL+"/auth/realms/test", "test-org", "test-user", "test-pass")

	// Create function client
	functionClient := newFunctionClient(server.URL, "test-ns", "/api/v1", authManager, &http.Client{})

	// Create update request
	updateReq := UpdateFunctionRequest{
		Image: "updated-image:latest",
	}

	// Test that update function returns error for 404 response
	ctx := context.Background()
	function, err := functionClient.UpdateFunction(ctx, "nonexistent-function", updateReq)

	// Should return an error for 404 response
	require.Error(t, err, "UpdateFunction should fail with 404 Not Found response")
	require.Nil(t, function, "Function should be nil on error")
	assert.Contains(t, err.Error(), "404", "Error should contain 404 status code")
}
