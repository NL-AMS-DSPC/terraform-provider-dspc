package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizationClient_CreateRole(t *testing.T) {
	tests := []struct {
		name           string
		roleName       string
		actions        []string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful creation",
			roleName:       "vm-operator",
			actions:        []string{"vm:CreateVM", "vm:ListVMs"},
			mockResponse:   map[string]string{"message": "role created successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "bad request",
			roleName:       "",
			actions:        []string{},
			mockResponse:   map[string]string{"error": "validation error"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "server error",
			roleName:       "vm-operator",
			actions:        []string{"vm:CreateVM"},
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.CreateRole(context.Background(), tt.roleName, tt.actions)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_GetRole(t *testing.T) {
	tests := []struct {
		name           string
		roleName       string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "successful get",
			roleName: "vm-operator",
			mockResponse: &Role{
				Name:    "vm-operator",
				Actions: []string{"vm:CreateVM", "vm:ListVMs"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			roleName:       "nonexistent-role",
			mockResponse:   map[string]string{"error": "role not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			role, err := client.GetRole(context.Background(), tt.roleName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.roleName, role.Name)
				assert.Equal(t, []string{"vm:CreateVM", "vm:ListVMs"}, role.Actions)
			}
		})
	}
}

func TestAuthorizationClient_ListRoles(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list",
			mockResponse: []RoleListItem{
				{Name: "vm-operator"},
				{Name: "storage-admin"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   []RoleListItem{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "server error",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			roles, err := client.ListRoles(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, roles, tt.expectedCount)
			}
		})
	}
}

func TestAuthorizationClient_UpdateRole(t *testing.T) {
	tests := []struct {
		name            string
		roleName        string
		actions         []string
		putStatusCode   int
		getResponse     interface{}
		getStatusCode   int
		expectError     bool
		expectedActions []string
	}{
		{
			name:          "successful update",
			roleName:      "vm-operator",
			actions:       []string{"vm:CreateVM", "vm:DeleteVM"},
			putStatusCode: http.StatusOK,
			getResponse: &Role{
				Name:    "vm-operator",
				Actions: []string{"vm:CreateVM", "vm:DeleteVM"},
			},
			getStatusCode:   http.StatusOK,
			expectError:     false,
			expectedActions: []string{"vm:CreateVM", "vm:DeleteVM"},
		},
		{
			name:          "put fails — role not found",
			roleName:      "nonexistent-role",
			actions:       []string{"vm:CreateVM"},
			putStatusCode: http.StatusNotFound,
			getResponse:   nil,
			getStatusCode: 0,
			expectError:   true,
		},
		{
			name:          "put succeeds but get fails",
			roleName:      "vm-operator",
			actions:       []string{"vm:CreateVM"},
			putStatusCode: http.StatusOK,
			getResponse:   map[string]string{"error": "internal server error"},
			getStatusCode: http.StatusInternalServerError,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				callCount++
				if r.Method == http.MethodPut {
					w.WriteHeader(tt.putStatusCode)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "role updated successfully"})
					return
				}
				w.WriteHeader(tt.getStatusCode)
				_ = json.NewEncoder(w).Encode(tt.getResponse)
			}))
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			role, err := client.UpdateRole(context.Background(), tt.roleName, tt.actions)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.roleName, role.Name)
				assert.Equal(t, tt.expectedActions, role.Actions)
				assert.Equal(t, 2, callCount)
			}
		})
	}
}

func TestAuthorizationClient_DeleteRole(t *testing.T) {
	tests := []struct {
		name           string
		roleName       string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"message": "role deleted successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			roleName:       "nonexistent-role",
			mockResponse:   map[string]string{"error": "role not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.DeleteRole(context.Background(), tt.roleName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_CreateGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful creation",
			groupName:      "platform-team",
			mockResponse:   map[string]string{"message": "group created successfully"},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "bad request",
			groupName:      "",
			mockResponse:   map[string]string{"error": "validation error"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "server error",
			groupName:      "platform-team",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.CreateGroup(context.Background(), tt.groupName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_GetGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful get",
			groupName:      "platform-team",
			mockResponse:   &Group{Name: "platform-team"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			groupName:      "nonexistent-group",
			mockResponse:   map[string]string{"error": "group not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			groupName:      "platform-team",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			group, err := client.GetGroup(context.Background(), tt.groupName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.groupName, group.Name)
			}
		})
	}
}

func TestAuthorizationClient_DeleteGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			groupName:      "platform-team",
			mockResponse:   map[string]string{"message": "group deleted successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			groupName:      "nonexistent-group",
			mockResponse:   map[string]string{"error": "group not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.DeleteGroup(context.Background(), tt.groupName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_AddUserToGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		userID         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful add",
			groupName:      "platform-team",
			userID:         "user-uuid-1234",
			mockResponse:   map[string]string{"message": "user added to group successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "group not found",
			groupName:      "nonexistent-group",
			userID:         "user-uuid-1234",
			mockResponse:   map[string]string{"error": "group not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			groupName:      "platform-team",
			userID:         "user-uuid-1234",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.AddUserToGroup(context.Background(), tt.groupName, tt.userID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_RemoveUserFromGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		userID         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful removal",
			groupName:      "platform-team",
			userID:         "user-uuid-1234",
			mockResponse:   map[string]string{"message": "user removed from group successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			groupName:      "nonexistent-group",
			userID:         "user-uuid-1234",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.RemoveUserFromGroup(context.Background(), tt.groupName, tt.userID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_AddRoleToGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		roleName       string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful add",
			groupName:      "platform-team",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"message": "role added to group successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "group not found",
			groupName:      "nonexistent-group",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"error": "group not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			groupName:      "platform-team",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.AddRoleToGroup(context.Background(), tt.groupName, tt.roleName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_RemoveRoleFromGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		roleName       string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful removal",
			groupName:      "platform-team",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"message": "role removed from group successfully"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			groupName:      "nonexistent-group",
			roleName:       "vm-operator",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			err := client.RemoveRoleFromGroup(context.Background(), tt.groupName, tt.roleName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizationClient_GetRolesForGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupName      string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name:           "successful list",
			groupName:      "platform-team",
			mockResponse:   []string{"vm-operator", "storage-admin"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			groupName:      "platform-team",
			mockResponse:   []string{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "group not found",
			groupName:      "nonexistent-group",
			mockResponse:   map[string]string{"error": "group not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Authorization

			roles, err := client.GetRolesForGroup(context.Background(), tt.groupName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, roles, tt.expectedCount)
			}
		})
	}
}

func TestAuthorizationClient_CreateRole_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		expectedPath := DefaultServiceConfig().Authorization.PathPrefix + "/v1/roles"
		assert.Equal(t, expectedPath, r.URL.Path)

		var req CreateRoleRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "vm-operator", req.Name)
		assert.Equal(t, []string{"vm:CreateVM", "vm:ListVMs"}, req.Actions)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "role created successfully"})
	}))
	defer server.Close()

	client := newTestAscClient(server.URL, authServer.URL).Authorization
	err := client.CreateRole(context.Background(), "vm-operator", []string{"vm:CreateVM", "vm:ListVMs"})
	assert.NoError(t, err)
}

func TestAuthorizationClient_UpdateRole_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			expectedPath := DefaultServiceConfig().Authorization.PathPrefix + "/v1/roles/vm-operator"
			assert.Equal(t, expectedPath, r.URL.Path)

			var req UpdateRoleRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, []string{"vm:CreateVM", "vm:DeleteVM"}, req.Actions)

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "role updated successfully"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&Role{Name: "vm-operator", Actions: []string{"vm:CreateVM", "vm:DeleteVM"}})
	}))
	defer server.Close()

	client := newTestAscClient(server.URL, authServer.URL).Authorization
	role, err := client.UpdateRole(context.Background(), "vm-operator", []string{"vm:CreateVM", "vm:DeleteVM"})
	assert.NoError(t, err)
	assert.Equal(t, "vm-operator", role.Name)
}

func TestAuthorizationClient_CreateGroup_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		expectedPath := DefaultServiceConfig().Authorization.PathPrefix + "/v1/groups"
		assert.Equal(t, expectedPath, r.URL.Path)

		var req CreateGroupRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "platform-team", req.GroupName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "group created successfully"})
	}))
	defer server.Close()

	client := newTestAscClient(server.URL, authServer.URL).Authorization
	err := client.CreateGroup(context.Background(), "platform-team")
	assert.NoError(t, err)
}
