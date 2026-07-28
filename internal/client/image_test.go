package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListImages(t *testing.T) {
	tests := map[string]struct {
		mockResponse   any
		mockStatusCode int
		expectedResult []ImageResponse
		expectError    bool
	}{
		"successful list": {
			mockResponse: []ImageResponse{
				{
					ID:                     "image-id",
					Name:                   "image-name",
					Family:                 "image-family",
					Distribution:           "image-distribution",
					Release:                "image-release",
					RequiresLicense:        false,
					SupportedArchitectures: []string{"arch1", "arch2"},
				},
			},
			mockStatusCode: http.StatusOK,
			expectedResult: []ImageResponse{
				{
					ID:                     "image-id",
					Name:                   "image-name",
					Family:                 "image-family",
					Distribution:           "image-distribution",
					Release:                "image-release",
					RequiresLicense:        false,
					SupportedArchitectures: []string{"arch1", "arch2"},
				},
			},
			expectError: false,
		},
		"server error": {
			mockResponse:   map[string]string{"error": "internal error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Images

			images, err := client.ListImages(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, images)
			}
		})
	}
}
