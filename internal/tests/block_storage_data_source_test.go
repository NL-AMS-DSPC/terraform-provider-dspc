package tests

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type BlockStorageDataSourceSuite struct {
	MockProvider
}

func TestBlockStorageDataSourcing(t *testing.T) {
	suite.Run(t, new(BlockStorageDataSourceSuite))
}

func (s *BlockStorageDataSourceSuite) TestBlockStorageDataSource_ClientError() {
	s.Handlers = map[string]func() MockResponse{
		"GET " + BuildTestPath("storage", "/blocks/test-block"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusInternalServerError,
				ResponseBody: "internal server error",
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
data "asc_block_storage" "test" {
name="test-block"
}`,
				ExpectError: regexp.MustCompile("internal server error"),
			},
		},
	})
}

func (s *BlockStorageDataSourceSuite) TestBlockStorageDataSource_GetBlock() {
	s.Handlers = map[string]func() MockResponse{
		"GET " + BuildTestPath("storage", "/blocks/test-block"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: struct {
					Name string
					Size string
				}{
					Name: "test-block",
					Size: "5Gi",
				},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: getProvider(s.Server.URL, s.AuthServer.URL,
					`
					data "asc_block_storage" "test" {
					  name="test-block"
					}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.asc_block_storage.test", "id", "test-block"),
					resource.TestCheckResourceAttr("data.asc_block_storage.test", "name", "test-block"),
					resource.TestCheckResourceAttr("data.asc_block_storage.test", "size", "5Gi"),
				),
			},
		},
	})
}
