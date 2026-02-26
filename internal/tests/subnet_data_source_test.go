package tests

import (
	"net/http"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type SubnetDataSourceSuite struct {
	MockProvider
}

func TestSubnetDataSource(t *testing.T) {
	suite.Run(t, new(SubnetDataSourceSuite))
}

func (s *SubnetDataSourceSuite) TestAccSubnetDataSource() {
	s.Handlers = MockResponses{
			"GET " + BuildTestPath("network", "/vpcs/test-vpc/subnets"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: []*client.Subnet{
					{Name: "public-subnet", CIDR: "10.0.0.0/25", Type: "public", VPCRef: "test-vpc", Status: "active"},
					{Name: "private-subnet", CIDR: "10.0.0.128/25", Type: "private", VPCRef: "test-vpc", Status: "active"},
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
data "dspc_subnets" "test" {
	vpc_name = "test-vpc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "vpc_name", "test-vpc"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.#", "2"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.name", "public-subnet"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.type", "public"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.name", "private-subnet"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.type", "private"),
				),
			},
		},
	})
}
