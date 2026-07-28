package client

import (
	"context"
	"net/http"
)

// ImageResponse represents the API response for an OS image.
type ImageResponse struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Family                 string   `json:"family"`
	Distribution           string   `json:"distribution"`
	Release                string   `json:"release"`
	RequiresLicense        bool     `json:"requiresLicense"`
	LicenseInfo            string   `json:"licenseInfo"`
	SupportedArchitectures []string `json:"supportedArchitectures"`
}

type imageClient struct {
	apiClient
}

// ListImages retrieves all available OS images
func (api *imageClient) ListImages(ctx context.Context) (images []ImageResponse, err error) {
	err = api.get(ctx, api.namespacedPath("/images"), &images)
	return
}

func newImageClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *imageClient {
	return &imageClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
