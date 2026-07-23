package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/suite"
)

type PostgreSQLDataSourceSuite struct {
	MockProvider
}

func TestPostgreSQLDataSource(t *testing.T) {
	suite.Run(t, new(PostgreSQLDataSourceSuite))
}

func (s *PostgreSQLDataSourceSuite) TestAccPostgreSQLDataSource() {
	s.Handlers = MockResponses{
		"GET " + BuildTestPath("mdb", "/databases/test-postgres"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: client.PostgreSQLInstance{
					Name:    "test-postgres",
					SkuSize: "gp-2",
					Version: client.DatabaseVersionPostgres17,
					VPCID:   "11111111-1111-1111-1111-111111111111",
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
data "asc_postgresql" "test" {
	name = "test-postgres"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.asc_postgresql.test", "name", "test-postgres"),
					resource.TestCheckResourceAttr("data.asc_postgresql.test", "sku_size", "gp-2"),
					resource.TestCheckResourceAttr("data.asc_postgresql.test", "version", "POSTGRES_17"),
					resource.TestCheckResourceAttr("data.asc_postgresql.test", "vpc_id", "11111111-1111-1111-1111-111111111111"),
				),
			},
		},
	})
}
