package client

import (
	"context"
	"fmt"
)

type DatabaseVersion string

const (
	DatabaseVersionMSSQL2025_17 DatabaseVersion = "MSSQL_2025_17"
	DatabaseVersionMSSQL2022_16 DatabaseVersion = "MSSQL_2022_16"
	DatabaseVersionMSSQL2019_15 DatabaseVersion = "MSSQL_2019_15"
	DatabaseVersionMSSQL2017_14 DatabaseVersion = "MSSQL_2017_14"
)

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MSSQLInstance struct {
	Name    string          `json:"name"`
	Size    string          `json:"size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

type CreateMSSQLInstanceRequest struct {
	Name    string          `json:"name"`
	Size    string          `json:"size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

type ListMSSQLInstancesResponse struct {
	Data []MSSQLInstance `json:"data"`
}

func (api *networkClient) CreateMSSQLInstance(ctx context.Context, req CreateMSSQLInstanceRequest) (instance *MSSQLInstance, err error) {
	err = api.post(ctx, api.namespacedPath("/databases"), req, &instance)
	return
}

func (api *networkClient) GetMSSQLInstance(ctx context.Context, instanceName string) (instance *MSSQLInstance, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/databases/%s", instanceName)), &instance)
	return
}

func (api *networkClient) ListMSSQLInstances(ctx context.Context) (instances *ListMSSQLInstancesResponse, err error) {
	err = api.get(ctx, api.namespacedPath("/databases"), &instances)
	return
}
