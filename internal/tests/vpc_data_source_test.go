package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
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
		"GET " + BuildTestPath("network", "/vpcs"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: []*client.VPC{
					{Name: "vpc-1", CIDR: "10.0.0.0/24", Status: "active", Subnets: []client.Subnet{{
						ID:        "s1-id",
						URN:       "s1-urn",
						Name:      "s1-name",
						CIDR:      "s1-cidr",
						Type:      "s1-type",
						VPCID:     "s1-vpc-id",
						Status:    "s1-status",
						LastError: "s1-last-error",
						Tags:      []client.Tag{{Key: "s1-t1-k", Value: "s1-t1-v"}},
					}}},
					{Name: "vpc-2", CIDR: "10.1.0.0/24", Status: "active", Tags: []client.Tag{{Key: "k1", Value: "v1"}}},
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
data "asc_vpcs" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.#", "2"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.name", "vpc-1"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.#", "1"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.id", "s1-id"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.urn", "s1-urn"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.name", "s1-name"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.cidr", "s1-cidr"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.type", "s1-type"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.vpc_id", "s1-vpc-id"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.status", "s1-status"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.last_error", "s1-last-error"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.tags.%", "1"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.subnets.0.tags.s1-t1-k", "s1-t1-v"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.0.tags.%", "0"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.1.name", "vpc-2"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.1.cidr", "10.1.0.0/24"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.1.subnets.#", "0"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.1.tags.%", "1"),
					resource.TestCheckResourceAttr("data.asc_vpcs.test", "vpcs.1.tags.k1", "v1"),
				),
			},
		},
	})
}
