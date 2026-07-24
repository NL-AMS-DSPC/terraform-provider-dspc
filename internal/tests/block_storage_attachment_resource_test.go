//#go:build tfacc

package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/suite"
)

type BlockStorageAttachmentResourceSuite struct {
	MockProvider
}

func TestAccBlockStorageAttachmentResourceSuite(t *testing.T) {
	suite.Run(t, new(BlockStorageAttachmentResourceSuite))
}

func (b *BlockStorageAttachmentResourceSuite) TestAccBlockStorageResource() {
	state := []*client.ListBlockAttachmentsForVMResponse{
		{
			Name:         "test-block",
			AttachedToVM: "test-vm",
		},
	}
	mock := MockResponse{
		ResponseCode: http.StatusOK,
		ResponseBody: state,
	}

	b.Handlers = MockResponses{
		"POST " + BuildTestPath("storage", "/blocks/test-block/attach/test-vm"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: client.CreateBlockAttachmentResponse{
					BlockName: "test-block",
					VMName:    "test-vm",
				},
			}
		},
		"GET " + BuildTestPath("vm", "/virtualmachines/test-vm/blocks"): func() MockResponse {
			return mock
		},
		"DELETE " + BuildTestPath("storage", "/blocks/test-block/attach/test-vm"): func() MockResponse {
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
resource "asc_block_storage_attachment" "test" {
	block_storage_name = "test-block"
	vm_name = "test-vm"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_block_storage_attachment.test", "id", "test-block:test-vm"),
					resource.TestCheckResourceAttr("asc_block_storage_attachment.test", "block_storage_name", "test-block"),
					resource.TestCheckResourceAttr("asc_block_storage_attachment.test", "vm_name", "test-vm"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "asc_block_storage_attachment.test",
				ImportState:       true,
				ImportStateId:     "test-block:test-vm",
				ImportStateVerify: true,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
