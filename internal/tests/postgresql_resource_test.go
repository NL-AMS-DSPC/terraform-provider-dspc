package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/suite"
)

type PostgreSQLResourceSuite struct {
	MockProvider
}

func TestPostgreSQLProvisioning(t *testing.T) {
	suite.Run(t, new(PostgreSQLResourceSuite))
}

func (s *PostgreSQLResourceSuite) TestAccPostgreSQLResource() {
	state := client.PostgreSQLInstance{
		Name:    "test-postgres",
		SkuSize: "gp-2",
		Version: client.DatabaseVersionPostgres17,
		VPCID:   "11111111-1111-1111-1111-111111111111",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("mdb", "/databases"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusCreated,
				ResponseBody: state,
			}
		},
		"GET " + BuildTestPath("mdb", "/databases/test-postgres"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: state,
			}
		},
		"DELETE " + BuildTestPath("mdb", "/databases/test-postgres"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusNoContent,
				ResponseBody: nil,
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "asc_postgresql" "test" {
	name    = "test-postgres"
	sku_size    = "gp-2"
	version = "POSTGRES_17"
	vpc_id     = "11111111-1111-1111-1111-111111111111"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_postgresql.test", "name", "test-postgres"),
					resource.TestCheckResourceAttr("asc_postgresql.test", "sku_size", "gp-2"),
					resource.TestCheckResourceAttr("asc_postgresql.test", "version", "POSTGRES_17"),
					resource.TestCheckResourceAttr("asc_postgresql.test", "vpc_id", "11111111-1111-1111-1111-111111111111"),
				),
			},
			// ImportState testing
			{
				ResourceName:  "asc_postgresql.test",
				ImportState:   true,
				ImportStateId: "test-postgres",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
