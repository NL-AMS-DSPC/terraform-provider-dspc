package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
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
resource "dspc_postgresql" "test" {
	name    = "test-postgres"
	size    = "1Gi"
	version = "POSTGRES_17"
	vpc     = "test-vpc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dspc_postgresql.test", "name", "test-postgres"),
					resource.TestCheckResourceAttr("dspc_postgresql.test", "size", "1Gi"),
					resource.TestCheckResourceAttr("dspc_postgresql.test", "version", "POSTGRES_17"),
					resource.TestCheckResourceAttr("dspc_postgresql.test", "vpc", "test-vpc"),
				),
			},
			// ImportState testing
			{
				ResourceName:  "dspc_postgresql.test",
				ImportState:   true,
				ImportStateId: "test-postgres",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
