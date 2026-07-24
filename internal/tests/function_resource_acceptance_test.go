package tests

import (
	"fmt"
	"net/http"
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
		"image":  "gcr.io/knative-samples/helloworld-go",
		"status": "ready",
	}

	updatedFunctionResponse := map[string]interface{}{
		"name":   "test-function",
		"image":  "gcr.io/knative-samples/updated-image:latest",
		"status": "ready",
	}

	// Mock API handlers
	s.Handlers = MockResponses{
		"POST " + BuildTestPath("function", "/v1/functions/"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: map[string]interface{}{
					"created": "test-function",
				},
			}
		},
		"GET " + BuildTestPath("function", "/v1/functions/test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]interface{}{"data": functionResponse},
			}
		},
		"PUT " + BuildTestPath("function", "/v1/functions/test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]interface{}{
					"updated": "test-function",
				},
			}
		},
		"DELETE " + BuildTestPath("function", "/v1/functions/test-function"): func() MockResponse {
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
resource "asc_function" "test" {
	name  = "test-function"
	image = "gcr.io/knative-samples/helloworld-go"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_function.test", "name", "test-function"),
					resource.TestCheckResourceAttr("asc_function.test", "image", "gcr.io/knative-samples/helloworld-go"),
					resource.TestCheckResourceAttr("asc_function.test", "id", "test-function"),
					resource.TestCheckResourceAttr("asc_function.test", "status", "ready"),
				),
			},
			// Update testing
			{
				PreConfig: func() {
					// Switch the GET handler to return the updated function after the PUT
					s.Handlers["GET "+BuildTestPath("function", "/v1/functions/test-function")] = func() MockResponse {
						return MockResponse{
							ResponseCode: http.StatusOK,
							ResponseBody: map[string]interface{}{"data": updatedFunctionResponse},
						}
					}
				},
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "asc_function" "test" {
	name  = "test-function"
	image = "gcr.io/knative-samples/updated-image:latest"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_function.test", "name", "test-function"),
					resource.TestCheckResourceAttr("asc_function.test", "image", "gcr.io/knative-samples/updated-image:latest"),
					resource.TestCheckResourceAttr("asc_function.test", "status", "ready"),
				),
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
				"POST " + BuildTestPath("function", "/v1/functions/"): func() MockResponse {
					return MockResponse{
						ResponseCode: http.StatusCreated,
						ResponseBody: map[string]interface{}{
							"created": tc.name,
						},
					}
				},
				"GET " + BuildTestPath("function", "/v1/functions/"+tc.name): func() MockResponse {
					return MockResponse{
						ResponseCode: http.StatusOK,
						ResponseBody: map[string]interface{}{"data": map[string]interface{}{
							"name":   tc.name,
							"image":  tc.image,
							"status": "ready",
						}},
					}
				},
				"DELETE " + BuildTestPath("function", "/v1/functions/"+tc.name): func() MockResponse {
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
resource "asc_function" "test" {
	name  = "%s"
	image = "%s"
}
`, tc.name, tc.image),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("asc_function.test", "name", tc.name),
							resource.TestCheckResourceAttr("asc_function.test", "image", tc.image),
							resource.TestCheckResourceAttr("asc_function.test", "id", tc.name),
							resource.TestCheckResourceAttr("asc_function.test", "status", "ready"),
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
		"image":  "gcr.io/knative-samples/helloworld-go",
		"status": "ready",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("function", "/v1/functions/"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: map[string]interface{}{
					"created": "import-test-function",
				},
			}
		},
		"GET " + BuildTestPath("function", "/v1/functions/import-test-function"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]interface{}{"data": functionResponse},
			}
		},
		"DELETE " + BuildTestPath("function", "/v1/functions/import-test-function"): func() MockResponse {
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
resource "asc_function" "test" {
	name  = "import-test-function"
	image = "gcr.io/knative-samples/helloworld-go"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_function.test", "name", "import-test-function"),
				),
			},
			// Import the resource
			{
				ResourceName:      "asc_function.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
