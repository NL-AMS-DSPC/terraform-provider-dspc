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

func TestBlockStorageAttachmentSuite(t *testing.T) {
	suite.Run(t, new(BlockStorageAttachmentSuite))
}

func (b *BlockStorageAttachmentSuite) TestAccBlockStorageDataSource() {
	state := []*client.ListBlockAttachmentsForVmResponse{
		{
			Name:         "block-test",
			AttachedToVM: "vm-test",
		},
	}
	mock := MockResponse{
		ResponseCode: http.StatusOK,
		ResponseBody: state,
	}

	b.Handlers = map[string]func() MockResponse{
		"GET /v1/namespaces/test-ns/virtualmachines/vm-test/pvcs": func() MockResponse {
			return mock
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
