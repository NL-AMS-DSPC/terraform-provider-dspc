package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/suite"
)

type BlockStorageResourceSuite struct {
	MockProvider
}

func TestBlockStorageProvisioning(t *testing.T) {
	suite.Run(t, new(BlockStorageResourceSuite))
}

func (b *BlockStorageResourceSuite) TestAccBlockStorageResource() {
	state := client.Block{
		Name: "test-block",
		Size: "10Gi",
	}

	mock := MockResponse{
		ResponseCode: http.StatusOK,
		ResponseBody: state,
	}

	b.Handlers = MockResponses{
		"POST " + BuildTestPath("storage", "/blocks"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: client.CreateBlockResponse{
					Created: "test-block",
				},
			}
		},
		"GET " + BuildTestPath("storage", "/blocks/test-block"): func() MockResponse {
			return mock
		},
		"DELETE " + BuildTestPath("storage", "/blocks/test-block"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]string{},
			}
		},
	}

	resource.Test(b.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read test
			{
				Config: TestProvider(b.Server.URL, b.AuthServer.URL) + `
resource "asc_block_storage" "test" {
	name = "test-block"
	size = "10Gi"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_block_storage.test", "name", "test-block"),
					resource.TestCheckResourceAttr("asc_block_storage.test", "size", "10Gi"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "asc_block_storage.test",
				ImportState:       true,
				ImportStateId:     "test-block",
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
