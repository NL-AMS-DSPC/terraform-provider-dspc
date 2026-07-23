package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/suite"
)

type FileStorageResourceSuite struct {
	MockProvider
}

func TestFileStorageProvisioning(t *testing.T) {
	suite.Run(t, new(FileStorageResourceSuite))
}

func (s *FileStorageResourceSuite) TestAccFileStorageResource() {
	state := client.FileStorage{
		Name:         "test-storage",
		Size:         "100Gi",
		Status:       "Ready",
		NFSMountPath: "/tenant-a/test-storage",
	}

	s.Handlers = MockResponses{
		"POST " + BuildTestPath("filestorage", "/file-storages"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: client.CreateFileStorageResponse{Created: "test-storage"},
			}
		},
		"GET " + BuildTestPath("filestorage", "/file-storages/test-storage"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: state,
			}
		},
		"DELETE " + BuildTestPath("filestorage", "/file-storages/test-storage"): func() MockResponse {
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
resource "asc_file_storage" "test" {
  name = "test-storage"
  size = "100Gi"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("asc_file_storage.test", "name", "test-storage"),
					resource.TestCheckResourceAttr("asc_file_storage.test", "size", "100Gi"),
					resource.TestCheckResourceAttr("asc_file_storage.test", "status", "Ready"),
					resource.TestCheckResourceAttr("asc_file_storage.test", "nfs_mount_path", "/tenant-a/test-storage"),
				),
			},
			{
				ResourceName:      "asc_file_storage.test",
				ImportState:       true,
				ImportStateId:     "test-storage",
				ImportStateVerify: true,
			},
		},
	})
}
