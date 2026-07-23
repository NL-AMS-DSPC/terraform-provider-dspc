package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
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
		ID:        "test-vpc-id",
		URN:       "test-vpc-urn",
		Name:      "test-vpc",
		CIDR:      "10.0.0.0/24",
		Status:    "active",
		LastError: "",
		Subnets: []client.Subnet{{
			ID:        "s1-id",
			URN:       "s1-urn",
			Name:      "s1-name",
			CIDR:      "s1-cidr",
			Type:      "s1-type",
			VPCID:     "s1-vpc-id",
			Status:    "s1-status",
			LastError: "s1-last-error",
			Tags:      []client.Tag{{Key: "s1-t1-k", Value: "s1-t1-v"}},
		}},
		Tags: []client.Tag{{Key: "k1", Value: "v1"}},
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
resource "asc_vpc" "test" {
	name = "test-vpc"
	subnets = [
		{
			name     = "s1-name"
			cidr     = "s1-cidr"
			type     = "s1-type"
			tags = {
				s1-t1-k = "s1-t1-v"
			}
		}
	]
	tags = {
		k1 = "v1"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_vpc.test", "id", "test-vpc-id"),
					resource.TestCheckResourceAttr("asc_vpc.test", "name", "test-vpc"),
					resource.TestCheckResourceAttr("asc_vpc.test", "cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("asc_vpc.test", "status", "active"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.#", "1"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.id", "s1-id"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.urn", "s1-urn"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.name", "s1-name"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.cidr", "s1-cidr"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.type", "s1-type"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.vpc_id", "s1-vpc-id"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.status", "s1-status"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.last_error", "s1-last-error"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.tags.%", "1"),
					resource.TestCheckResourceAttr("asc_vpc.test", "subnets.0.tags.s1-t1-k", "s1-t1-v"),
					resource.TestCheckResourceAttr("asc_vpc.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("asc_vpc.test", "tags.k1", "v1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "asc_vpc.test",
				ImportState:       true,
				ImportStateId:     "test-vpc",
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
