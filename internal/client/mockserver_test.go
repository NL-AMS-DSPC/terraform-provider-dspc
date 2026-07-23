package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

const namespace = "test-ns"

// mockResponse allows the mock HTTP server to mock multiple responses depending on the request path and method
type mockResponse struct {
	method     string
	path       string
	statusCode int
	response   interface{}
}

// newMockRouteServer returns a new mock HTTP server that allows multiple mock responses
func newMockRouteServer(pathPrefix string, responses []mockResponse) *httptest.Server {
	r := chi.NewRouter()

	namespacePath := fmt.Sprintf("%s/v1/namespaces/%s", pathPrefix, namespace)

	for _, response := range responses {
		r.MethodFunc(response.method, namespacePath+response.path, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(response.statusCode)
			_ = json.NewEncoder(w).Encode(response.response)
		})
	}

	return httptest.NewServer(r)
}

// newMockServer returns a simple mock HTTP server that returns a single response
func newMockServer(statusCode int, response interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)
	}))
}

// newTestAscClient creates a new DSPC client configured for testing
func newTestAscClient(endpoint, authURL string) *AscClient {
	return NewAscClient(endpoint, namespace, "test-client-id", "test-client-secret", authURL, "test-realm", 30)
}
