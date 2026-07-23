package group

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

// mockRoleAssignClient implements RoleResourceClient with function fields for test control.
type mockRoleAssignClient struct {
	addRoleToGroup      func(ctx context.Context, groupName, roleName string) error
	removeRoleFromGroup func(ctx context.Context, groupName, roleName string) error
	getRolesForGroup    func(ctx context.Context, groupName string) ([]string, error)
}

func (m *mockRoleAssignClient) AddRoleToGroup(ctx context.Context, groupName, roleName string) error {
	return m.addRoleToGroup(ctx, groupName, roleName)
}

func (m *mockRoleAssignClient) RemoveRoleFromGroup(ctx context.Context, groupName, roleName string) error {
	return m.removeRoleFromGroup(ctx, groupName, roleName)
}

func (m *mockRoleAssignClient) GetRolesForGroup(ctx context.Context, groupName string) ([]string, error) {
	return m.getRolesForGroup(ctx, groupName)
}

var roleAssignObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"group_name": tftypes.String,
		"role_name":  tftypes.String,
	},
}

func getRoleAssignSchema(t *testing.T, r *RoleResource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func makeRoleAssignRaw(id tftypes.Value, groupName, roleName string) tftypes.Value {
	return tftypes.NewValue(roleAssignObjectType, map[string]tftypes.Value{
		"id":         id,
		"group_name": tftypes.NewValue(tftypes.String, groupName),
		"role_name":  tftypes.NewValue(tftypes.String, roleName),
	})
}

func nullRoleAssignRaw() tftypes.Value {
	return tftypes.NewValue(roleAssignObjectType, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, nil),
		"group_name": tftypes.NewValue(tftypes.String, nil),
		"role_name":  tftypes.NewValue(tftypes.String, nil),
	})
}

func TestRoleResource_Create(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		roleName    string
		clientErr   error
		expectError bool
	}{
		{
			name:      "successful creation",
			groupName: "test-group",
			roleName:  "admin",
		},
		{
			name:        "client error",
			groupName:   "test-group",
			roleName:    "admin",
			clientErr:   errors.New("API error 409: assignment already exists"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &RoleResource{
				client: &mockRoleAssignClient{
					addRoleToGroup: func(_ context.Context, groupName, roleName string) error {
						if groupName != tt.groupName {
							t.Errorf("AddRoleToGroup: got groupName %q, want %q", groupName, tt.groupName)
						}
						if roleName != tt.roleName {
							t.Errorf("AddRoleToGroup: got roleName %q, want %q", roleName, tt.roleName)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getRoleAssignSchema(t, r)
			raw := makeRoleAssignRaw(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tt.groupName, tt.roleName)

			req := resource.CreateRequest{
				Plan: tfsdk.Plan{Schema: schResp.Schema, Raw: raw},
			}
			resp := &resource.CreateResponse{
				State: tfsdk.State{Schema: schResp.Schema},
			}

			r.Create(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
			}
			var model RoleResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			wantID := tt.groupName + ":" + tt.roleName
			if model.ID.ValueString() != wantID {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), wantID)
			}
			if model.GroupName.ValueString() != tt.groupName {
				t.Errorf("GroupName: got %q, want %q", model.GroupName.ValueString(), tt.groupName)
			}
			if model.RoleName.ValueString() != tt.roleName {
				t.Errorf("RoleName: got %q, want %q", model.RoleName.ValueString(), tt.roleName)
			}
		})
	}
}

func TestRoleResource_Read(t *testing.T) {
	tests := []struct {
		name             string
		groupName        string
		roleName         string
		getRolesForGroup func(ctx context.Context, groupName string) ([]string, error)
		expectError      bool
		expectGone       bool
	}{
		{
			name:      "role still assigned",
			groupName: "test-group",
			roleName:  "admin",
			getRolesForGroup: func(_ context.Context, _ string) ([]string, error) {
				return []string{"viewer", "admin"}, nil
			},
		},
		{
			name:      "role no longer assigned — removed from state",
			groupName: "test-group",
			roleName:  "admin",
			getRolesForGroup: func(_ context.Context, _ string) ([]string, error) {
				return []string{"viewer"}, nil
			},
			expectGone: true,
		},
		{
			name:      "group not found — removed from state",
			groupName: "gone-group",
			roleName:  "admin",
			getRolesForGroup: func(_ context.Context, _ string) ([]string, error) {
				return nil, errors.New("resource not found")
			},
			expectGone: true,
		},
		{
			name:      "API 404 — removed from state",
			groupName: "gone-group",
			roleName:  "admin",
			getRolesForGroup: func(_ context.Context, _ string) ([]string, error) {
				return nil, errors.New("API error 404: not found")
			},
			expectGone: true,
		},
		{
			name:      "client error",
			groupName: "test-group",
			roleName:  "admin",
			getRolesForGroup: func(_ context.Context, _ string) ([]string, error) {
				return nil, errors.New("connection refused")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &RoleResource{
				client: &mockRoleAssignClient{
					getRolesForGroup: tt.getRolesForGroup,
				},
			}

			schResp := getRoleAssignSchema(t, r)
			raw := makeRoleAssignRaw(
				tftypes.NewValue(tftypes.String, tt.groupName+":"+tt.roleName),
				tt.groupName, tt.roleName,
			)

			req := resource.ReadRequest{
				State: tfsdk.State{Schema: schResp.Schema, Raw: raw},
			}
			resp := &resource.ReadResponse{
				State: tfsdk.State{Schema: schResp.Schema, Raw: raw},
			}

			r.Read(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
				return
			}
			if tt.expectGone {
				if !resp.State.Raw.IsNull() {
					t.Error("expected state to be removed, but it was not null")
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
			}
			var model RoleResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.GroupName.ValueString() != tt.groupName {
				t.Errorf("GroupName: got %q, want %q", model.GroupName.ValueString(), tt.groupName)
			}
			if model.RoleName.ValueString() != tt.roleName {
				t.Errorf("RoleName: got %q, want %q", model.RoleName.ValueString(), tt.roleName)
			}
		})
	}
}

func TestRoleResource_Update(t *testing.T) {
	r := &RoleResource{}
	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected Update to always return a diagnostics error, got none")
	}
}

func TestRoleResource_Delete(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		roleName    string
		clientErr   error
		expectError bool
	}{
		{
			name:      "successful deletion",
			groupName: "test-group",
			roleName:  "admin",
		},
		{
			name:      "assignment not found — silently removed from state",
			groupName: "test-group",
			roleName:  "admin",
			clientErr: errors.New("resource not found"),
		},
		{
			name:        "API error",
			groupName:   "test-group",
			roleName:    "admin",
			clientErr:   errors.New("API error 500: internal server error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &RoleResource{
				client: &mockRoleAssignClient{
					removeRoleFromGroup: func(_ context.Context, groupName, roleName string) error {
						if groupName != tt.groupName {
							t.Errorf("RemoveRoleFromGroup: got groupName %q, want %q", groupName, tt.groupName)
						}
						if roleName != tt.roleName {
							t.Errorf("RemoveRoleFromGroup: got roleName %q, want %q", roleName, tt.roleName)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getRoleAssignSchema(t, r)
			raw := makeRoleAssignRaw(
				tftypes.NewValue(tftypes.String, tt.groupName+":"+tt.roleName),
				tt.groupName, tt.roleName,
			)

			req := resource.DeleteRequest{
				State: tfsdk.State{Schema: schResp.Schema, Raw: raw},
			}
			resp := &resource.DeleteResponse{}

			r.Delete(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("unexpected diagnostics error: %s", resp.Diagnostics)
				}
			}
		})
	}
}

func TestRoleResource_ImportState(t *testing.T) {
	tests := []struct {
		name        string
		importID    string
		wantGroup   string
		wantRole    string
		expectError bool
	}{
		{
			name:      "valid import ID",
			importID:  "my-group:admin",
			wantGroup: "my-group",
			wantRole:  "admin",
		},
		{
			name:        "missing separator",
			importID:    "my-group",
			expectError: true,
		},
		{
			name:        "empty group name",
			importID:    ":admin",
			expectError: true,
		},
		{
			name:        "empty role name",
			importID:    "my-group:",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &RoleResource{}
			schResp := getRoleAssignSchema(t, r)

			req := resource.ImportStateRequest{ID: tt.importID}
			resp := &resource.ImportStateResponse{
				State: tfsdk.State{Schema: schResp.Schema, Raw: nullRoleAssignRaw()},
			}

			r.ImportState(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
			}
			var model RoleResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.GroupName.ValueString() != tt.wantGroup {
				t.Errorf("GroupName: got %q, want %q", model.GroupName.ValueString(), tt.wantGroup)
			}
			if model.RoleName.ValueString() != tt.wantRole {
				t.Errorf("RoleName: got %q, want %q", model.RoleName.ValueString(), tt.wantRole)
			}
		})
	}
}

func TestRoleResource_Metadata(t *testing.T) {
	r := &RoleResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "asc"}, resp)
	if resp.TypeName != "asc_group_role" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "asc_group_role")
	}
}

func TestRoleResource_Schema(t *testing.T) {
	r := &RoleResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "group_name", "role_name"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing attribute %q", attr)
		}
	}
}

func TestRoleResource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{
			name:         "nil provider data — skipped",
			providerData: nil,
		},
		{
			name:         "invalid provider data type",
			providerData: "not-a-client",
			expectError:  true,
		},
		{
			name:         "valid DspcClient",
			providerData: client.NewDspcClient("http://localhost", "test-ns", "test-user", "test-pass", "http://auth.example.com", "test-org", 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RoleResource{}
			req := resource.ConfigureRequest{ProviderData: tt.providerData}
			resp := &resource.ConfigureResponse{}
			r.Configure(context.Background(), req, resp)
			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("unexpected diagnostics error: %s", resp.Diagnostics)
				}
			}
		})
	}
}

func TestNewRoleResource(t *testing.T) {
	if NewRoleResource() == nil {
		t.Error("NewRoleResource returned nil")
	}
}
