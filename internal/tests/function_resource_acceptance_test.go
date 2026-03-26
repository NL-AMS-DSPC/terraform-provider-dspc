package tests

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type FunctionResourceSuite struct {
	MockProvider
}

func TestAccFunctionResource(t *testing.T) {
	suite.Run(t, new(FunctionResourceSuite))
}

func (s *FunctionResourceSuite) TestAccFunctionResource() {
	// Mock function creation response
	functionResponse := map[string]interface{}{
		"name":   "test-function",
		"status": "ready",
	}

	// Mock API handlers
	s.Handlers = MockResponses{
		"POST " + BuildTestPath("vm", "/virtualmachines/"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: map[string]interface{}{
					"created": "test-function",
				},
			}
		},
		"GET " + BuildTestPath("vm", "/virtualmachines/test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: functionResponse,
			}
		},
		"DELETE " + BuildTestPath("vm", "/virtualmachines/test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]interface{}{
					"deleted": "test-function",
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "dspc_function" "test" {
	namespace = "test-ns"
	name      = "test-function"
	image     = "gcr.io/knative-samples/helloworld-go"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify flattened schema attributes are accessible
					resource.TestCheckResourceAttr("dspc_function.test", "namespace", "test-ns"),
					resource.TestCheckResourceAttr("dspc_function.test", "name", "test-function"),
					resource.TestCheckResourceAttr("dspc_function.test", "image", "gcr.io/knative-samples/helloworld-go"),
					resource.TestCheckResourceAttr("dspc_function.test", "id", "test-function"),
					resource.TestCheckResourceAttr("dspc_function.test", "status", "ready"),
				),
			},
			// Update testing (should fail as updates are not supported)
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "dspc_function" "test" {
	namespace = "test-ns"
	name      = "test-function"
	image     = "gcr.io/knative-samples/updated-image:latest"
}
`,
				ExpectError: regexp.MustCompile("Update not supported"),
			},
		},
	})
}

func (s *FunctionResourceSuite) TestAccFunctionResourceWithDifferentImages() {
	// Test different image configurations
	testCases := []struct {
		name  string
		image string
	}{
		{
			name:  "test-function-1",
			image: "nginx:alpine",
		},
		{
			name:  "test-function-2",
			image: "ghcr.io/my-org/my-app:v1.0.0",
		},
	}

	for _, tc := range testCases {
		s.T().Run(fmt.Sprintf("function_%s", tc.name), func(t *testing.T) {
			// Setup mock responses for this specific function
			s.Handlers = MockResponses{
				"POST " + BuildTestPath("vm", "/virtualmachines/"): func() MockResponse {
					return MockResponse{
						ResponseCode: http.StatusCreated,
						ResponseBody: map[string]interface{}{
							"created": tc.name,
						},
					}
				},
				"GET " + BuildTestPath("vm", "/virtualmachines/"+tc.name): func() MockResponse {
					return MockResponse{
						ResponseCode: http.StatusOK,
						ResponseBody: map[string]interface{}{
							"name":   tc.name,
							"status": "ready",
						},
					}
				},
				"DELETE " + BuildTestPath("vm", "/virtualmachines/"+tc.name): func() MockResponse {
					return MockResponse{
						ResponseCode: http.StatusOK,
						ResponseBody: map[string]interface{}{
							"deleted": tc.name,
						},
					}
				},
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: TestProvider(s.Server.URL, s.AuthServer.URL) + fmt.Sprintf(`
resource "dspc_function" "test" {
	namespace = "test-ns"
	name      = "%s"
	image     = "%s"
}
`, tc.name, tc.image),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("dspc_function.test", "name", tc.name),
							resource.TestCheckResourceAttr("dspc_function.test", "image", tc.image),
							resource.TestCheckResourceAttr("dspc_function.test", "id", tc.name),
							resource.TestCheckResourceAttr("dspc_function.test", "status", "ready"),
						),
					},
				},
			})
		})
	}
}

func (s *FunctionResourceSuite) TestAccFunctionResourceImport() {
	// Test import functionality
	functionResponse := map[string]interface{}{
		"name":   "import-test-function",
		"status": "ready",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("vm", "/virtualmachines/"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: map[string]interface{}{
					"created": "import-test-function",
				},
			}
		},
		"GET " + BuildTestPath("vm", "/virtualmachines/import-test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: functionResponse,
			}
		},
		"DELETE " + BuildTestPath("vm", "/virtualmachines/import-test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]interface{}{
					"deleted": "import-test-function",
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the resource
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "dspc_function" "test" {
	namespace = "test-ns"
	name      = "import-test-function"
	image     = "gcr.io/knative-samples/helloworld-go"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dspc_function.test", "name", "import-test-function"),
				),
			},
			// Import the resource
			{
				ResourceName:      "dspc_function.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
