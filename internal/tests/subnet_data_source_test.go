package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
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
					{ID: "s1", URN: "s1-urn", Name: "public-subnet", CIDR: "10.0.0.0/25", Type: "public", VPCID: "test-vpc", Status: "active", Tags: []client.Tag{{Key: "k1", Value: "v1"}}},
					{ID: "s2", URN: "s2-urn", Name: "private-subnet", CIDR: "10.0.0.128/25", Type: "private", VPCID: "test-vpc", Status: "active", LastError: "foo"},
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
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.id", "s1"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.urn", "s1-urn"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.name", "public-subnet"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.cidr", "10.0.0.0/25"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.type", "public"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.status", "active"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.last_error", ""),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.tags.%", "1"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.0.tags.k1", "v1"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.id", "s2"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.urn", "s2-urn"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.name", "private-subnet"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.cidr", "10.0.0.128/25"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.type", "private"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.status", "active"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.last_error", "foo"),
					resource.TestCheckResourceAttr("data.dspc_subnets.test", "subnets.1.tags.%", "0"),
				),
			},
		},
	})
}
