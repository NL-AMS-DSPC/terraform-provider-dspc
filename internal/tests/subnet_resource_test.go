package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/suite"
)

type SubnetResourceSuite struct {
	MockProvider
}

func TestSubnetProvisioning(t *testing.T) {
	suite.Run(t, new(SubnetResourceSuite))
}

func (s *SubnetResourceSuite) TestAccSubnetResource() {
	createdSubnet := client.Subnet{
		Name:   "test-subnet",
		CIDR:   "10.0.0.0/25",
		Type:   "public",
		VPCID:  "test-vpc",
		Status: "active",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("network", "/vpcs/test-vpc/subnets"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: createdSubnet,
			}
		},
		"GET " + BuildTestPath("network", "/vpcs/test-vpc/subnets"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: []*client.Subnet{&createdSubnet},
			}
		},
		"DELETE " + BuildTestPath("network", "/vpcs/test-vpc/subnets/test-subnet"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]string{"deleted": "test-subnet"},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "dspc_subnet" "test" {
	name     = "test-subnet"
	vpc_name = "test-vpc"
	vpc_id   = "test-vpc-id"
	cidr     = "10.0.0.0/25"
	type     = "public"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dspc_subnet.test", "name", "test-subnet"),
					resource.TestCheckResourceAttr("dspc_subnet.test", "vpc_name", "test-vpc"),
					resource.TestCheckResourceAttr("dspc_subnet.test", "vpc_id", "test-vpc-id"),
					resource.TestCheckResourceAttr("dspc_subnet.test", "cidr", "10.0.0.0/25"),
					resource.TestCheckResourceAttr("dspc_subnet.test", "type", "public"),
					resource.TestCheckResourceAttr("dspc_subnet.test", "status", "active"),
					resource.TestCheckResourceAttr("dspc_subnet.test", "id", "test-vpc:test-subnet"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dspc_subnet.test",
				ImportState:       true,
				ImportStateId:     "test-vpc:test-subnet",
				ImportStateVerify: true,
				// vpc_id is not returned by the list-subnets API used for Read, so it
				// cannot be reconstructed on import.
				ImportStateVerifyIgnore: []string{"vpc_id"},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
