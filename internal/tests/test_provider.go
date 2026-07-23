// nolint:all
package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/provider"
	"github.com/stretchr/testify/suite"
)

const (
	providerConfig = `
provider "asc" {
	endpoint  = "%s"
	auth_url  = "%s"
	namespace = "test-ns"
	org       = "test-realm"
	username  = "test-client-id"
	password  = "test-client-secret"
  	timeout   = 60
}
`
)

// TestAccProtoV6ProviderFactories is a map of provider factories used for Terraform acceptance testing with ProtoV6.
var (
	//
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"asc": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

// TestProvider generates the provider configuration string using the given endpoint and auth URLs.
func TestProvider(endpointURL, authURL string) string {
	return fmt.Sprintf(providerConfig, endpointURL, authURL)
}

func getProvider(baseURL, authURL string, modules ...string) string {
	terraformModules := fmt.Sprintf(`
	provider "asc" {
		endpoint  = "%s"
		auth_url  = "%s"
		namespace = "test-ns"
		org       = "test-realm"
		username  = "test-client-id"
		password  = "test-client-secret"
  		timeout   = 60
	}`, baseURL, authURL)

	for _, m := range modules {
		terraformModules += m
	}

	return terraformModules
}

// MockResponse represents a simulated HTTP response for testing or mocking purposes.
// It contains the HTTP status code and the response body.
type MockResponse struct {
	ResponseCode int
	ResponseBody any
}

// MockProvider is a test utility struct that provides an httptest.Server and configurable mock response handlers.
type MockProvider struct {
	suite.Suite

	Server     *httptest.Server
	AuthServer *httptest.Server
	Handlers   MockResponses
}

// MockResponses is a type alias for a map of string keys to functions returning a MockResponse. It defines mock HTTP handlers.
type MockResponses = map[string]func() MockResponse

// SetupTest initializes a test HTTP server and sets up request handlers for mocking HTTP responses.
func (s *MockProvider) SetupTest() {
	// Setup mock Keycloak auth server
	s.AuthServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Mock Keycloak token endpoint
		if req.URL.Path == "/realms/test-realm/protocol/openid-connect/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-jwt",
				"expires_in":   3600,
				"token_type":   "Bearer",
			})
			return
		}
		s.T().Fatalf("Unexpected auth request: %s", req.URL.Path)
	}))

	// Setup main API server
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
		println("request path", requestPath)
		handler, ok := s.Handlers[requestPath]
		if !ok {
			s.T().Fatalf("Invalid request done on path: %s", requestPath)
		}
		resp := handler()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.ResponseCode)
		_ = json.NewEncoder(w).Encode(resp.ResponseBody)
	}))
}

// TearDownTest stops the test HTTP server and cleans up resources used by the MockProvider.
func (s *MockProvider) TearDownTest() {
	s.Server.Close()
	s.AuthServer.Close()
}
