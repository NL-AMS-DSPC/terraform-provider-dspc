package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/suite"
)

type FileStorageDataSourceSuite struct {
	MockProvider
}

func TestFileStorageDataSource(t *testing.T) {
	suite.Run(t, new(FileStorageDataSourceSuite))
}

func (s *FileStorageDataSourceSuite) TestAccFileStorageDataSource() {
	state := client.FileStorage{
		Name:         "test-storage",
		Size:         "50Gi",
		Status:       "Ready",
		NFSMountPath: "/tenant-a/test-storage",
	}

	s.Handlers = MockResponses{
		"GET " + BuildTestPath("filestorage", "/file-storages/test-storage"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: state,
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
data "asc_file_storage" "test" {
  name = "test-storage"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.asc_file_storage.test", "name", "test-storage"),
					resource.TestCheckResourceAttr("data.asc_file_storage.test", "size", "50Gi"),
					resource.TestCheckResourceAttr("data.asc_file_storage.test", "status", "Ready"),
					resource.TestCheckResourceAttr("data.asc_file_storage.test", "nfs_mount_path", "/tenant-a/test-storage"),
				),
			},
		},
	})
}
