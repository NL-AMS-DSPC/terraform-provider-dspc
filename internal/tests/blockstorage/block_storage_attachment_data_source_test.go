package blockstorage

import (
	"net/http"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/provider"
	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/tests"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dspc": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

type BlockStorageAttachmentSuite struct {
	tests.MockProvider
}

func TestMySuite(t *testing.T) {
	suite.Run(t, new(BlockStorageAttachmentSuite))
}

func (b *BlockStorageAttachmentSuite) TestAccBlockStorageDataSource() {
	b.Response = tests.MockResponse{
		ResponseCode: http.StatusOK,
		ResponseBody: []client.ListBlockAttachmentsForVmResponse{
			{
				Name:         "block-test",
				AttachedToVM: "vm-test",
			},
		},
	}
	resource.Test(b.T(), resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tests.TestProvider(b.Server.URL) + `
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
