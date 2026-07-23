package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
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
		ID:     "test-subnet-id",
		URN:    "test-subnet-urn",
		Name:   "test-subnet",
		CIDR:   "10.0.0.0/25",
		Type:   "public",
		VPCID:  "test-vpc-id",
		Status: "active",
		Tags: []client.Tag{
			{
				Key:   "k1",
				Value: "v1",
			},
		},
		LastError: "",
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
resource "asc_subnet" "test" {
	name     = "test-subnet"
	vpc_name = "test-vpc"
	vpc_id   = "test-vpc-id"
	cidr     = "10.0.0.0/25"
	type     = "public"
	tags = {
		k1 = "v1"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_subnet.test", "id", "test-subnet-id"),
					resource.TestCheckResourceAttr("asc_subnet.test", "urn", "test-subnet-urn"),
					resource.TestCheckResourceAttr("asc_subnet.test", "name", "test-subnet"),
					resource.TestCheckResourceAttr("asc_subnet.test", "vpc_name", "test-vpc"),
					resource.TestCheckResourceAttr("asc_subnet.test", "vpc_id", "test-vpc-id"),
					resource.TestCheckResourceAttr("asc_subnet.test", "cidr", "10.0.0.0/25"),
					resource.TestCheckResourceAttr("asc_subnet.test", "type", "public"),
					resource.TestCheckResourceAttr("asc_subnet.test", "status", "active"),
					resource.TestCheckResourceAttr("asc_subnet.test", "last_error", ""),
					resource.TestCheckResourceAttr("asc_subnet.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("asc_subnet.test", "tags.k1", "v1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "asc_subnet.test",
				ImportState:       true,
				ImportStateId:     "test-vpc:test-subnet",
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
