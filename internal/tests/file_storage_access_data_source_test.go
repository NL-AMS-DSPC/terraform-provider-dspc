package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/suite"
)

type FileStorageAccessDataSourceSuite struct {
	MockProvider
}

func TestFileStorageAccessDataSource(t *testing.T) {
	suite.Run(t, new(FileStorageAccessDataSourceSuite))
}

func (s *FileStorageAccessDataSourceSuite) TestAccFileStorageAccessDataSource() {
	accessState := client.FileStorageAccess{
		FileStorageName: "test-storage",
		TargetType:      "Container",
		TargetName:      "my-container",
	}

	s.Handlers = MockResponses{
		"GET " + BuildTestPath("filestorage", "/file-storages/test-storage/access/Container/my-container"): func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: accessState,
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestProvider(s.Server.URL, s.AuthServer.URL) + `
data "asc_file_storage_access" "test" {
  file_storage_name = "test-storage"
  target_type       = "Container"
  target_name       = "my-container"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.asc_file_storage_access.test", "file_storage_name", "test-storage"),
					resource.TestCheckResourceAttr("data.asc_file_storage_access.test", "target_type", "Container"),
					resource.TestCheckResourceAttr("data.asc_file_storage_access.test", "target_name", "my-container"),
					resource.TestCheckResourceAttr("data.asc_file_storage_access.test", "id", "test-storage:Container:my-container"),
				),
			},
		},
	})
}
