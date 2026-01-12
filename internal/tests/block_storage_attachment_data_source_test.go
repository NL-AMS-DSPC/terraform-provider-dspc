package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	providerConfig = `
provider "dspc" {
	endpoint = "%s"
	namespace= "test-ns"
  	timeout  = 60
  	api_key  = "your-api-key-here" 
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dspc": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

func TestAccBlockStorageDataSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := []client.ListBlockAttachmentsForVmResponse{
			{
				Name:         "block-test",
				AttachedToVM: "vm-test",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(providerConfig, server.URL) + `
data "dspc_block_storage_attachment" "test" {
	block_storage_name = "block-test"
	vm_name = "vm-test"	
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "block_storage_name", "block-test"),
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "vm_name", "vm-test"),
					resource.TestCheckResourceAttr("data.dspc_block_storage_attachment.test", "id", "block-test-vm-test"),
				),
			},
		},
	})
}
