package client

import "net/http"

type managedDatabaseClient struct {
	apiClient
}

func newManagedDatabaseClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *managedDatabaseClient {
	return &managedDatabaseClient{
		apiClient: newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
