package tests

import (
	"net/http"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type BlockStorageAttachmentSuite struct {
	MockProvider
}

func TestMySuite(t *testing.T) {
	suite.Run(t, new(BlockStorageAttachmentSuite))
}

func (b *BlockStorageAttachmentSuite) TestAccBlockStorageDataSource() {
	b.Response = MockResponse{
		ResponseCode: http.StatusOK,
		ResponseBody: []client.ListBlockAttachmentsForVmResponse{
			{
				Name:         "block-test",
				AttachedToVM: "vm-test",
			},
		},
	}
	resource.Test(b.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(b.Server.URL) + `
data "dspc_block_storage_attachment" "test" {
	block_storage_name = "block-test"
	vm_name = "vm-test"	
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "id", "block-test-vm-test"),
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "block_storage_name", "block-test"),
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "vm_name", "vm-test"),
				),
			},
		},
	})
}
