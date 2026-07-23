package role

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// mockResourceClient implements ResourceClient with function fields for test control.
type mockResourceClient struct {
	createRole func(ctx context.Context, name string, actions []string) error
	getRole    func(ctx context.Context, name string) (*client.Role, error)
	updateRole func(ctx context.Context, name string, actions []string) (*client.Role, error)
	deleteRole func(ctx context.Context, name string) error
}

func (m *mockResourceClient) CreateRole(ctx context.Context, name string, actions []string) error {
	return m.createRole(ctx, name, actions)
}

func (m *mockResourceClient) GetRole(ctx context.Context, name string) (*client.Role, error) {
	return m.getRole(ctx, name)
}

func (m *mockResourceClient) UpdateRole(ctx context.Context, name string, actions []string) (*client.Role, error) {
	return m.updateRole(ctx, name, actions)
}

func (m *mockResourceClient) DeleteRole(ctx context.Context, name string) error {
	return m.deleteRole(ctx, name)
}

var roleObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":      tftypes.String,
		"name":    tftypes.String,
		"actions": tftypes.List{ElementType: tftypes.String},
	},
}

func getResourceSchema(t *testing.T, r *Resource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func makePlanRaw(id tftypes.Value, name string, actions []string) tftypes.Value {
	actionVals := make([]tftypes.Value, len(actions))
	for i, a := range actions {
		actionVals[i] = tftypes.NewValue(tftypes.String, a)
	}
	return tftypes.NewValue(roleObjectType, map[string]tftypes.Value{
		"id":      id,
		"name":    tftypes.NewValue(tftypes.String, name),
		"actions": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, actionVals),
	})
}

// nullStateRaw is a fully-typed object with all-null attributes, suitable for
// initialising an ImportStateResponse or ReadResponse where the state has no prior value.
func nullStateRaw() tftypes.Value {
	return tftypes.NewValue(roleObjectType, map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, nil),
		"name":    tftypes.NewValue(tftypes.String, nil),
		"actions": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	})
}

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		actions     []string
		clientErr   error
		expectError bool
	}{
		{
			name:     "successful creation",
			roleName: "test-role",
			actions:  []string{"vm:CreateVM", "vm:ListVMs"},
		},
		{
			name:     "successful creation with no actions",
			roleName: "empty-role",
			actions:  []string{},
		},
		{
			name:        "client error",
			roleName:    "bad-role",
			actions:     []string{"vm:CreateVM"},
			clientErr:   errors.New("API error 409: role already exists"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockResourceClient{
					createRole: func(_ context.Context, name string, _ []string) error {
						if name != tt.roleName {
							t.Errorf("CreateRole: got name %q, want %q", name, tt.roleName)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getResourceSchema(t, r)
			raw := makePlanRaw(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tt.roleName, tt.actions)

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
			var model ResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.ID.ValueString() != tt.roleName {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), tt.roleName)
			}
			if model.Name.ValueString() != tt.roleName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.roleName)
			}
		})
	}
}

func TestResource_Read(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		getRole     func(ctx context.Context, name string) (*client.Role, error)
		expectError bool
		expectGone  bool
	}{
		{
			name:     "successful read",
			roleName: "test-role",
			getRole: func(_ context.Context, name string) (*client.Role, error) {
				return &client.Role{Name: name, Actions: []string{"vm:CreateVM"}}, nil
			},
		},
		{
			name:     "role not found removes from state",
			roleName: "gone-role",
			getRole: func(_ context.Context, _ string) (*client.Role, error) {
				return nil, errors.New("resource not found")
			},
			expectGone: true,
		},
		{
			name:     "API 404 removes from state",
			roleName: "gone-role",
			getRole: func(_ context.Context, _ string) (*client.Role, error) {
				return nil, errors.New("API error 404: not found")
			},
			expectGone: true,
		},
		{
			name:     "client error",
			roleName: "test-role",
			getRole: func(_ context.Context, _ string) (*client.Role, error) {
				return nil, errors.New("connection refused")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockResourceClient{
					getRole: tt.getRole,
				},
			}

			schResp := getResourceSchema(t, r)
			raw := makePlanRaw(tftypes.NewValue(tftypes.String, tt.roleName), tt.roleName, []string{"vm:CreateVM"})

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
			var model ResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.ID.ValueString() != tt.roleName {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), tt.roleName)
			}
			if model.Name.ValueString() != tt.roleName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.roleName)
			}
		})
	}
}

func TestResource_Update(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		newActions  []string
		updatedRole *client.Role
		clientErr   error
		expectError bool
	}{
		{
			name:       "successful update",
			roleName:   "test-role",
			newActions: []string{"vm:CreateVM", "uam:ListUsers"},
			updatedRole: &client.Role{
				Name:    "test-role",
				Actions: []string{"vm:CreateVM", "uam:ListUsers"},
			},
		},
		{
			name:        "client error",
			roleName:    "test-role",
			newActions:  []string{"vm:CreateVM"},
			clientErr:   errors.New("API error 500: internal server error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockResourceClient{
					updateRole: func(_ context.Context, name string, _ []string) (*client.Role, error) {
						if name != tt.roleName {
							t.Errorf("UpdateRole: got name %q, want %q", name, tt.roleName)
						}
						if tt.clientErr != nil {
							return nil, tt.clientErr
						}
						return tt.updatedRole, nil
					},
				},
			}

			schResp := getResourceSchema(t, r)
			raw := makePlanRaw(tftypes.NewValue(tftypes.String, tt.roleName), tt.roleName, tt.newActions)

			req := resource.UpdateRequest{
				Plan:  tfsdk.Plan{Schema: schResp.Schema, Raw: raw},
				State: tfsdk.State{Schema: schResp.Schema, Raw: raw},
			}
			resp := &resource.UpdateResponse{
				State: tfsdk.State{Schema: schResp.Schema},
			}

			r.Update(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
			}
			var model ResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.ID.ValueString() != tt.roleName {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), tt.roleName)
			}
			if model.Name.ValueString() != tt.roleName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.roleName)
			}
			var gotActions []string
			if diags := model.Actions.ElementsAs(ctx, &gotActions, false); diags.HasError() {
				t.Fatalf("failed to read actions: %s", diags)
			}
			if fmt.Sprint(gotActions) != fmt.Sprint(tt.updatedRole.Actions) {
				t.Errorf("Actions: got %v, want %v", gotActions, tt.updatedRole.Actions)
			}
		})
	}
}

func TestResource_Delete(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		clientErr   error
		expectError bool
	}{
		{
			name:     "successful deletion",
			roleName: "test-role",
		},
		{
			name:      "role not found — silently removed from state",
			roleName:  "gone-role",
			clientErr: errors.New("resource not found"),
		},
		{
			name:        "API error",
			roleName:    "test-role",
			clientErr:   errors.New("API error 500: internal server error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockResourceClient{
					deleteRole: func(_ context.Context, name string) error {
						if name != tt.roleName {
							t.Errorf("DeleteRole: got name %q, want %q", name, tt.roleName)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getResourceSchema(t, r)
			raw := makePlanRaw(tftypes.NewValue(tftypes.String, tt.roleName), tt.roleName, []string{"vm:CreateVM"})

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

func TestResource_ImportState(t *testing.T) {
	ctx := context.Background()
	r := &Resource{}
	schResp := getResourceSchema(t, r)

	req := resource.ImportStateRequest{ID: "my-role"}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: schResp.Schema, Raw: nullStateRaw()},
	}

	r.ImportState(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
	}
	var model ResourceModel
	if diags := resp.State.Get(ctx, &model); diags.HasError() {
		t.Fatalf("failed to read state: %s", diags)
	}
	if model.Name.ValueString() != "my-role" {
		t.Errorf("Name: got %q, want %q", model.Name.ValueString(), "my-role")
	}
}

func TestResource_Metadata(t *testing.T) {
	r := &Resource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "asc"}, resp)
	if resp.TypeName != "asc_role" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "asc_role")
	}
}

func TestResource_Schema(t *testing.T) {
	r := &Resource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "name", "actions"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing attribute %q", attr)
		}
	}
}

func TestResource_Configure(t *testing.T) {
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
			name:         "valid AscClient",
			providerData: client.NewAscClient("http://localhost", "test-ns", "test-user", "test-pass", "http://auth.example.com", "test-org", 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resource{}
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
