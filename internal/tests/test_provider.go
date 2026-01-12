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

var (
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dspc": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

func TestProvider(url string) string {
	return fmt.Sprintf(providerConfig, url)
}

type MockResponse struct {
	ResponseCode int
	ResponseBody any
}

type MockProvider struct {
	suite.Suite

	Server   *httptest.Server
	Handlers MockResponses
}

type MockResponses = map[string]func() MockResponse

func (s *MockProvider) SetupTest() {
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestPath := fmt.Sprintf("%s %s", req.Method, req.URL.Path)
		resp := s.Handlers[requestPath]()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.ResponseCode)
		_ = json.NewEncoder(w).Encode(resp.ResponseBody)
	}))
}

func (s *MockProvider) TearDownTest() {
	s.Server.Close()
}
