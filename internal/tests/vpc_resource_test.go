package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/suite"
)

type VPCResourceSuite struct {
	MockProvider
}

func TestVPCProvisioning(t *testing.T) {
	suite.Run(t, new(VPCResourceSuite))
}

func (s *VPCResourceSuite) TestAccVPCResource() {
	state := client.VPC{
		Name:   "test-vpc",
		CIDR:   "10.0.0.0/24",
		Status: "active",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("network", "/vpcs"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: state,
			}
		},
		"GET " + BuildTestPath("network", "/vpcs/test-vpc"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: state,
			}
		},
		"DELETE " + BuildTestPath("network", "/vpcs/test-vpc"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]string{"deleted": "test-vpc"},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "dspc_vpc" "test" {
	name = "test-vpc"
	cidr = "10.0.0.0/24"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dspc_vpc.test", "name", "test-vpc"),
					resource.TestCheckResourceAttr("dspc_vpc.test", "cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("dspc_vpc.test", "status", "active"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dspc_vpc.test",
				ImportState:       true,
				ImportStateId:     "test-vpc",
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
