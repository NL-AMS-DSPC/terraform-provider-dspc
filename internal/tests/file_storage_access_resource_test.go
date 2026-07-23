package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/suite"
)

type FileStorageAccessResourceSuite struct {
	MockProvider
}

func TestFileStorageAccessProvisioning(t *testing.T) {
	suite.Run(t, new(FileStorageAccessResourceSuite))
}

func (s *FileStorageAccessResourceSuite) TestAccFileStorageAccessResource() {
	accessState := client.FileStorageAccess{
		FileStorageName: "test-storage",
		TargetType:      "VirtualMachine",
		TargetName:      "my-vm",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("filestorage", "/file-storages/test-storage/access"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: accessState,
			}
		},
		"GET " + BuildTestPath("filestorage", "/file-storages/test-storage/access/VirtualMachine/my-vm"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: accessState,
			}
		},
		"DELETE " + BuildTestPath("filestorage", "/file-storages/test-storage/access/VirtualMachine/my-vm"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: map[string]string{},
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
resource "asc_file_storage_access" "test" {
  file_storage_name = "test-storage"
  target_type       = "VirtualMachine"
  target_name       = "my-vm"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_file_storage_access.test", "file_storage_name", "test-storage"),
					resource.TestCheckResourceAttr("asc_file_storage_access.test", "target_type", "VirtualMachine"),
					resource.TestCheckResourceAttr("asc_file_storage_access.test", "target_name", "my-vm"),
					resource.TestCheckResourceAttr("asc_file_storage_access.test", "id", "test-storage:VirtualMachine:my-vm"),
				),
			},
			{
				ResourceName:      "asc_file_storage_access.test",
				ImportState:       true,
				ImportStateId:     "test-storage:VirtualMachine:my-vm",
				ImportStateVerify: true,
			},
		},
	})
}
