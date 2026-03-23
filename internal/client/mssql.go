package client

import (
	"context"
	"fmt"
)

// DatabaseVersion represents the version of the MSSQL database instance.
type DatabaseVersion string

const (
	// DatabaseVersionMSSQL2025_17 represents Microsoft SQL Server 2025 version 17.
	DatabaseVersionMSSQL2025_17 DatabaseVersion = "MSSQL_2025_17"
	// DatabaseVersionMSSQL2022_16 represents Microsoft SQL Server 2022 version 16.
	DatabaseVersionMSSQL2022_16 DatabaseVersion = "MSSQL_2022_16"
	// DatabaseVersionMSSQL2019_15 represents Microsoft SQL Server 2019 version 15.
	DatabaseVersionMSSQL2019_15 DatabaseVersion = "MSSQL_2019_15"
	// DatabaseVersionMSSQL2017_14 represents Microsoft SQL Server 2017 version 14.
	DatabaseVersionMSSQL2017_14 DatabaseVersion = "MSSQL_2017_14"
)

// Tag represents a key-value pair used for tagging MSSQL instances.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MSSQLInstance represents a Microsoft SQL Server database instance with its properties.
type MSSQLInstance struct {
	Name    string          `json:"name"`
	Size    string          `json:"size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

// CreateMSSQLInstanceRequest represents the request payload for creating a new MSSQL instance.
type CreateMSSQLInstanceRequest struct {
	Name    string          `json:"name"`
	Size    string          `json:"size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

// ListMSSQLInstancesResponse represents the response payload for listing MSSQL instances, containing a slice of MSSQLInstance.
type ListMSSQLInstancesResponse struct {
	Data []MSSQLInstance `json:"data"`
}

// CreateMSSQLInstance creates a new MSSQL instance with the specified properties and returns the created instance.
func (api *networkClient) CreateMSSQLInstance(ctx context.Context, req CreateMSSQLInstanceRequest) (instance *MSSQLInstance, err error) {
	err = api.post(ctx, api.namespacedPath("/databases"), req, &instance)
	return
}

// GetMSSQLInstance retrieves the details of a specific MSSQL instance by its name and returns the instance information.
func (api *networkClient) GetMSSQLInstance(ctx context.Context, instanceName string) (instance *MSSQLInstance, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/databases/%s", instanceName)), &instance)
	return
}

// ListMSSQLInstances retrieves a list of all MSSQL instances and returns them in a ListMSSQLInstancesResponse struct.
func (api *networkClient) ListMSSQLInstances(ctx context.Context) (instances *ListMSSQLInstancesResponse, err error) {
	err = api.get(ctx, api.namespacedPath("/databases"), &instances)
	return
}
