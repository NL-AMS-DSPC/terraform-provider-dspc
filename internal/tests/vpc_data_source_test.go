package tests

import (
	"net/http"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type VPCDataSourceSuite struct {
	MockProvider
}

func TestVPCDataSource(t *testing.T) {
	suite.Run(t, new(VPCDataSourceSuite))
}

func (s *VPCDataSourceSuite) TestAccVPCDataSource() {
	s.Handlers = MockResponses{
		"GET /v1/namespaces/test-ns/vpcs": func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: []*client.VPC{
					{Name: "vpc-1", CIDR: "10.0.0.0/24", Status: "active"},
					{Name: "vpc-2", CIDR: "10.1.0.0/24", Status: "active"},
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL) + `
data "dspc_vpcs" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dspc_vpcs.test", "vpcs.#", "2"),
					resource.TestCheckResourceAttr("data.dspc_vpcs.test", "vpcs.0.name", "vpc-1"),
					resource.TestCheckResourceAttr("data.dspc_vpcs.test", "vpcs.0.cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("data.dspc_vpcs.test", "vpcs.1.name", "vpc-2"),
					resource.TestCheckResourceAttr("data.dspc_vpcs.test", "vpcs.1.cidr", "10.1.0.0/24"),
				),
			},
		},
	})
}
