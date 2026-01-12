package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

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
	Response MockResponse
}

func (s *MockProvider) SetupTest() {
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(s.Response.ResponseBody)
	}))
}

func (s *MockProvider) TearDownTest() {
	s.Server.Close()
}
