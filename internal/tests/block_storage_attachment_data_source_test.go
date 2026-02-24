// Package tests contains test suites for the DSPC provider.
package tests

import (
	"net/http"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type BlockStorageAttachmentDatasourceSuite struct {
	MockProvider
}

func TestBlockStorageAttachmentDatasourceSuite(t *testing.T) {
	suite.Run(t, new(BlockStorageAttachmentDatasourceSuite))
}

func (b *BlockStorageAttachmentDatasourceSuite) TestAccBlockStorageDataSource() {
	state := []*client.ListBlockAttachmentsForVMResponse{
		{
			Name:         "block-test",
			AttachedToVM: "vm-test",
		},
	}
	mock := MockResponse{
		ResponseCode: http.StatusOK,
		ResponseBody: state,
	}

	b.Handlers = MockResponses{
		"GET /api/vm/v1/namespaces/test-ns/virtualmachines/vm-test/blocks": func() MockResponse {
			return mock
		},
	}

	resource.Test(b.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(b.Server.URL, b.AuthServer.URL) + `
data "dspc_block_storage_attachment" "test" {
	block_storage_name = "block-test"
	vm_name = "vm-test"	
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "id", "block-test:vm-test"),
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "block_storage_name", "block-test"),
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "vm_name", "vm-test"),
				),
			},
		},
	})
}
