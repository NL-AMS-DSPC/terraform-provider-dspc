package client

import (
	"context"
	"fmt"
	"net/http"
)

// Role represents a role and its associated permissions.
type Role struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions,omitempty"`
}

// RoleListItem represents a role entry in the list-roles response.
type RoleListItem struct {
	Name string `json:"name"`
}

// CreateRoleRequest contains the parameters for creating a new role.
type CreateRoleRequest struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

// UpdateRoleRequest contains the parameters for updating an existing role.
type UpdateRoleRequest struct {
	Actions []string `json:"actions"`
}

type authorizationClient struct {
	apiClient
}

// CreateRole creates a new role with the specified permissions.
func (api *authorizationClient) CreateRole(ctx context.Context, name string, actions []string) error {
	req := CreateRoleRequest{Name: name, Actions: actions}
	return api.post(ctx, "/v1/roles", req, nil)
}

// UpdateRole updates the actions of an existing role and returns the updated role.
func (api *authorizationClient) UpdateRole(ctx context.Context, name string, actions []string) (*Role, error) {
	req := UpdateRoleRequest{Actions: actions}
	if err := api.put(ctx, fmt.Sprintf("/v1/roles/%s", name), req, nil); err != nil {
		return nil, err
	}
	return api.GetRole(ctx, name)
}

// DeleteRole deletes a role by name.
func (api *authorizationClient) DeleteRole(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/roles/%s", name))
}

// GetRole retrieves a role and its associated permissions by name.
func (api *authorizationClient) GetRole(ctx context.Context, name string) (*Role, error) {
	var role Role
	err := api.get(ctx, fmt.Sprintf("/v1/roles/%s", name), &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListRoles retrieves all roles.
func (api *authorizationClient) ListRoles(ctx context.Context) ([]RoleListItem, error) {
	var roles []RoleListItem
	err := api.get(ctx, "/v1/roles", &roles)
	return roles, err
}

// Group represents an authorization group.
type Group struct {
	Name string `json:"name"`
}

// CreateGroupRequest contains the parameters for creating a new group.
type CreateGroupRequest struct {
	GroupName string `json:"groupName"`
}

// CreateGroup creates a new group.
func (api *authorizationClient) CreateGroup(ctx context.Context, name string) error {
	req := CreateGroupRequest{GroupName: name}
	return api.post(ctx, "/v1/groups", req, nil)
}

// GetGroup retrieves a group by name.
func (api *authorizationClient) GetGroup(ctx context.Context, name string) (*Group, error) {
	var g Group
	err := api.get(ctx, fmt.Sprintf("/v1/groups/%s", name), &g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// DeleteGroup deletes a group by name.
func (api *authorizationClient) DeleteGroup(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/groups/%s", name))
}

// AddUserToGroup adds a user to a group.
func (api *authorizationClient) AddUserToGroup(ctx context.Context, groupName, userID string) error {
	return api.post(ctx, fmt.Sprintf("/v1/groups/%s/members/%s", groupName, userID), nil, nil)
}

// RemoveUserFromGroup removes a user from a group.
func (api *authorizationClient) RemoveUserFromGroup(ctx context.Context, groupName, userID string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/groups/%s/members/%s", groupName, userID))
}

// AddRoleToGroup adds a role to a group.
func (api *authorizationClient) AddRoleToGroup(ctx context.Context, groupName, roleName string) error {
	return api.post(ctx, fmt.Sprintf("/v1/groups/%s/roles/%s", groupName, roleName), nil, nil)
}

// RemoveRoleFromGroup removes a role from a group.
func (api *authorizationClient) RemoveRoleFromGroup(ctx context.Context, groupName, roleName string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/groups/%s/roles/%s", groupName, roleName))
}

// GetRolesForGroup retrieves all roles assigned to a group.
func (api *authorizationClient) GetRolesForGroup(ctx context.Context, groupName string) ([]string, error) {
	var roles []string
	err := api.get(ctx, fmt.Sprintf("/v1/groups/%s/roles", groupName), &roles)
	return roles, err
}

func newAuthorizationClient(endpoint, pathPrefix string, authMgr *authManager, httpClient *http.Client) *authorizationClient {
	// namespace is not used by the authorization service — pass empty string
	return &authorizationClient{
		newAPIClient(endpoint, "", pathPrefix, authMgr, httpClient),
	}
}
