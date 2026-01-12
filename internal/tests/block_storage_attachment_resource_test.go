//#go:build tfacc

package tests

import (
	"net/http"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type BlockStorageAttachmentResourceSuite struct {
	MockProvider
}

func TestAccBlockStorageAttachmentResourceSuite(t *testing.T) {
	suite.Run(t, new(BlockStorageAttachmentResourceSuite))
}

func (b *BlockStorageAttachmentResourceSuite) TestAccBlockStorageResource() {
	state := []*client.ListBlockAttachmentsForVmResponse{
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
		"POST /v1/namespaces/test-ns/pvcs/test-block/attach/test-vm": func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: client.CreateBlockAttachmentResponse{
					BlockName: "test-block",
					VMName:    "test-vm",
				},
			}
		},
		"GET /v1/namespaces/test-ns/virtualmachines/test-vm/pvcs": func() MockResponse {
			return mock
		},
		"DELETE /v1/namespaces/test-ns/pvcs/test-block/attach/test-vm": func() MockResponse {
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
				Config: TestProvider(b.Server.URL) + `
resource "dspc_block_storage_attachment" "test" {
	block_storage_name = "test-block"
	vm_name = "test-vm"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dspc_block_storage_attachment.test", "id", "test-block-test-vm"),
					resource.TestCheckResourceAttr("dspc_block_storage_attachment.test", "block_storage_name", "test-block"),
					resource.TestCheckResourceAttr("dspc_block_storage_attachment.test", "vm_name", "test-vm"),
				),
			},
		},
	})
}
