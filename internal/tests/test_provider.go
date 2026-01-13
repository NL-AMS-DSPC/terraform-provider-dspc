package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/suite"
)

const (
	providerConfig = `
provider "dspc" {
	endpoint = "%s"
	namespace= "test-ns"
  	timeout  = 60
  	api_key  = "your-api-key-here" 
}
`
)

// TestAccProtoV6ProviderFactories is a map of provider factories used for Terraform acceptance testing with ProtoV6.
var (
	//
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dspc": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

// TestProvider generates the provider configuration string using the given endpoint URL.
func TestProvider(url string) string {
	return fmt.Sprintf(providerConfig, url)
}

func getProvider(baseURL string, modules ...string) string {
	terraformModules := fmt.Sprintf(`
	provider "dspc" {
		endpoint = "%s"
		namespace= "test-ns"
  		timeout  = 60
  		api_key  = "your-api-key-here" 
	}`, baseURL)

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

	Server   *httptest.Server
	Handlers MockResponses
}

// MockResponses is a type alias for a map of string keys to functions returning a MockResponse. It defines mock HTTP handlers.
type MockResponses = map[string]func() MockResponse

// SetupTest initializes a test HTTP server and sets up request handlers for mocking HTTP responses.
func (s *MockProvider) SetupTest() {
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
}
