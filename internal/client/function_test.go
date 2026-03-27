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

const testFunctionPath = "/api/v1/v1/functions/test-function"

func TestFunctionClient_DeleteFunction_204Response(t *testing.T) {
	// Create a test server for both auth and API requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth requests (POST to /auth/realms/*)
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/auth/realms/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		// Handle function delete requests that return 204 No Content
		if r.Method == http.MethodDelete && r.URL.Path == testFunctionPath {
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create auth manager with test server URL
	authManager := newAuthManager(&http.Client{}, server.URL+"/auth/realms/test", "test-org", "test-user", "test-pass")

	// Create function client
	functionClient := newFunctionClient(server.URL, "test-ns", "/api/v1", authManager, &http.Client{})

	// Test that delete function succeeds with 204 response
	ctx := context.Background()
	err := functionClient.DeleteFunction(ctx, "test-function")

	// Should not return an error for 204 response
	assert.NoError(t, err, "DeleteFunction should succeed with 204 No Content response")
}

func TestFunctionClient_DeleteFunction_200Response(t *testing.T) {
	// Create a test server for both auth and API requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth requests (POST to /auth/realms/*)
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/auth/realms/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		// Handle function delete requests that return 200 OK
		if r.Method == http.MethodDelete && r.URL.Path == testFunctionPath {
			w.WriteHeader(http.StatusOK) // 200 OK
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create auth manager with test server URL
	authManager := newAuthManager(&http.Client{}, server.URL+"/auth/realms/test", "test-org", "test-user", "test-pass")

	// Create function client
	functionClient := newFunctionClient(server.URL, "test-ns", "/api/v1", authManager, &http.Client{})

	// Test that delete function succeeds with 200 response
	ctx := context.Background()
	err := functionClient.DeleteFunction(ctx, "test-function")

	// Should not return an error for 200 response
	assert.NoError(t, err, "DeleteFunction should succeed with 200 OK response")
}

func TestFunctionClient_DeleteFunction_404Response(t *testing.T) {
	// Create a test server for both auth and API requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth requests (POST to /auth/realms/*)
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/auth/realms/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"test-token","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		// Handle all other requests as 404
		w.WriteHeader(http.StatusNotFound) // 404 Not Found
		_, _ = w.Write([]byte("Function not found"))
	}))
	defer server.Close()

	// Create auth manager with test server URL
	authManager := newAuthManager(&http.Client{}, server.URL+"/auth/realms/test", "test-org", "test-user", "test-pass")

	// Create function client
	functionClient := newFunctionClient(server.URL, "test-ns", "/api/v1", authManager, &http.Client{})

	// Test that delete function returns error for 404 response
	ctx := context.Background()
	err := functionClient.DeleteFunction(ctx, "test-function")

	// Should return an error for 404 response
	require.Error(t, err, "DeleteFunction should fail with 404 Not Found response")
	assert.ErrorIs(t, err, ErrResourceNotFound, "Error should be ErrResourceNotFound for 404 response")
}
