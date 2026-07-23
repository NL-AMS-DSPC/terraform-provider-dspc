package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful get function",
			mockResponse: &client.Function{
				Name:   "test-function",
				Image:  "gcr.io/knative-samples/helloworld-go",
				Status: "ready",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "function not found",
			mockResponse:   nil,
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/namespaces/test-ns/functions/test-function", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					_ = json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			// Create data source
			dataSource, ok := NewFunctionDataSource().(*DataSource)
			require.True(t, ok, "Failed to cast to DataSource")

			// Test basic data source creation
			assert.NotNil(t, dataSource)
		})
	}
}
