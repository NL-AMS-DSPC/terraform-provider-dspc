package client

import (
	"context"
	"fmt"
)

const (
	// DatabaseVersionPostgres15 represents PostgreSQL version 15.
	DatabaseVersionPostgres15 DatabaseVersion = "POSTGRES_15"
	// DatabaseVersionPostgres16 represents PostgreSQL version 16.
	DatabaseVersionPostgres16 DatabaseVersion = "POSTGRES_16"
	// DatabaseVersionPostgres17 represents PostgreSQL version 17.
	DatabaseVersionPostgres17 DatabaseVersion = "POSTGRES_17"
	// DatabaseVersionPostgres18 represents PostgreSQL version 18.
	DatabaseVersionPostgres18 DatabaseVersion = "POSTGRES_18"
)

// PostgreSQLInstance represents a PostgreSQL database instance with its properties.
type PostgreSQLInstance struct {
	Name    string          `json:"name"`
	Size    string          `json:"storage_size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

// CreatePostgreSQLInstanceRequest represents the request payload for creating a new PostgreSQL instance.
type CreatePostgreSQLInstanceRequest struct {
	Name    string          `json:"name"`
	Size    string          `json:"storage_size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

// UpdatePostgreSQLInstanceRequest represents the request payload for updating an existing PostgreSQL instance.
type UpdatePostgreSQLInstanceRequest struct {
	Name    string          `json:"name"`
	Size    string          `json:"storage_size"`
	Version DatabaseVersion `json:"version"`
	VPC     string          `json:"vpc"`
	Tags    []Tag           `json:"tags,omitempty"`
}

// ListPostgreSQLInstancesResponse represents the response payload for listing PostgreSQL instances.
type ListPostgreSQLInstancesResponse struct {
	Data []PostgreSQLInstance `json:"data"`
}

// CreatePostgreSQLInstance creates a new PostgreSQL instance with the specified properties and returns the created instance.
func (api *managedDatabaseClient) CreatePostgreSQLInstance(ctx context.Context, req CreatePostgreSQLInstanceRequest) (instance *PostgreSQLInstance, err error) {
	err = api.post(ctx, "/v1/databases", req, &instance)
	return
}

// GetPostgreSQLInstance retrieves the details of a specific PostgreSQL instance by its name.
func (api *managedDatabaseClient) GetPostgreSQLInstance(ctx context.Context, instanceName string) (instance *PostgreSQLInstance, err error) {
	err = api.get(ctx, fmt.Sprintf("/v1/databases/%s", instanceName), &instance)
	return
}

// ListPostgreSQLInstances retrieves a list of all PostgreSQL instances.
func (api *managedDatabaseClient) ListPostgreSQLInstances(ctx context.Context) (instances *ListPostgreSQLInstancesResponse, err error) {
	err = api.get(ctx, "/v1/databases", &instances)
	return
}

// UpdatePostgreSQLInstance updates an existing PostgreSQL instance with the specified properties.
func (api *managedDatabaseClient) UpdatePostgreSQLInstance(ctx context.Context, instanceName string, req UpdatePostgreSQLInstanceRequest) (instance *PostgreSQLInstance, err error) {
	err = api.put(ctx, fmt.Sprintf("/v1/databases/%s", instanceName), req, &instance)
	return
}

// DeletePostgreSQLInstance deletes the PostgreSQL instance with the given name.
func (api *managedDatabaseClient) DeletePostgreSQLInstance(ctx context.Context, instanceName string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/databases/%s", instanceName))
}
