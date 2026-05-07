package client

import "net/http"

type managedDatabaseClient struct {
	apiClient
}

func newManagedDatabaseClient(endpoint, pathPrefix string, authMgr *authManager, httpClient *http.Client) *managedDatabaseClient {
	return &managedDatabaseClient{
		apiClient: newAPIClient(endpoint, "", pathPrefix, authMgr, httpClient),
	}
}
