package group

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// mockMemberResourceClient implements MemberResourceClient with function fields for test control.
type mockMemberResourceClient struct {
	addUserToGroup    func(ctx context.Context, groupName, userID string) error
	removeUserFromGroup func(ctx context.Context, groupName, userID string) error
}

func (m *mockMemberResourceClient) AddUserToGroup(ctx context.Context, groupName, userID string) error {
	return m.addUserToGroup(ctx, groupName, userID)
}

func (m *mockMemberResourceClient) RemoveUserFromGroup(ctx context.Context, groupName, userID string) error {
	return m.removeUserFromGroup(ctx, groupName, userID)
}

var memberObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":         tftypes.String,
		"group_name": tftypes.String,
		"user_id":    tftypes.String,
	},
}

func getMemberSchema(t *testing.T, r *MemberResource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func makeMemberRaw(id tftypes.Value, groupName, userID string) tftypes.Value {
	return tftypes.NewValue(memberObjectType, map[string]tftypes.Value{
		"id":         id,
		"group_name": tftypes.NewValue(tftypes.String, groupName),
		"user_id":    tftypes.NewValue(tftypes.String, userID),
	})
}

func nullMemberRaw() tftypes.Value {
	return tftypes.NewValue(memberObjectType, map[string]tftypes.Value{
		"id":         tftypes.NewValue(tftypes.String, nil),
		"group_name": tftypes.NewValue(tftypes.String, nil),
		"user_id":    tftypes.NewValue(tftypes.String, nil),
	})
}

func TestMemberResource_Create(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		userID      string
		clientErr   error
		expectError bool
	}{
		{
			name:      "successful creation",
			groupName: "test-group",
			userID:    "user-123",
		},
		{
			name:        "client error",
			groupName:   "test-group",
			userID:      "user-123",
			clientErr:   errors.New("API error 409: membership already exists"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &MemberResource{
				client: &mockMemberResourceClient{
					addUserToGroup: func(_ context.Context, groupName, userID string) error {
						if groupName != tt.groupName {
							t.Errorf("AddUserToGroup: got groupName %q, want %q", groupName, tt.groupName)
						}
						if userID != tt.userID {
							t.Errorf("AddUserToGroup: got userID %q, want %q", userID, tt.userID)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getMemberSchema(t, r)
			raw := makeMemberRaw(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tt.groupName, tt.userID)

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
			var model MemberResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			wantID := tt.groupName + ":" + tt.userID
			if model.ID.ValueString() != wantID {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), wantID)
			}
			if model.GroupName.ValueString() != tt.groupName {
				t.Errorf("GroupName: got %q, want %q", model.GroupName.ValueString(), tt.groupName)
			}
			if model.UserID.ValueString() != tt.userID {
				t.Errorf("UserID: got %q, want %q", model.UserID.ValueString(), tt.userID)
			}
		})
	}
}

func TestMemberResource_Read(t *testing.T) {
	// Read is a no-op; it must not modify state or return diagnostics.
	r := &MemberResource{}
	schResp := getMemberSchema(t, r)
	raw := makeMemberRaw(tftypes.NewValue(tftypes.String, "test-group:user-123"), "test-group", "user-123")

	req := resource.ReadRequest{
		State: tfsdk.State{Schema: schResp.Schema, Raw: raw},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{Schema: schResp.Schema, Raw: raw},
	}

	r.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected diagnostics error: %s", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Error("Read should preserve state, but state was removed")
	}
}

func TestMemberResource_Update(t *testing.T) {
	r := &MemberResource{}
	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected Update to always return a diagnostics error, got none")
	}
}

func TestMemberResource_Delete(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		userID      string
		clientErr   error
		expectError bool
	}{
		{
			name:      "successful deletion",
			groupName: "test-group",
			userID:    "user-123",
		},
		{
			name:      "membership not found — silently removed from state",
			groupName: "test-group",
			userID:    "user-123",
			clientErr: errors.New("resource not found"),
		},
		{
			name:        "API error",
			groupName:   "test-group",
			userID:      "user-123",
			clientErr:   errors.New("API error 500: internal server error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &MemberResource{
				client: &mockMemberResourceClient{
					removeUserFromGroup: func(_ context.Context, groupName, userID string) error {
						if groupName != tt.groupName {
							t.Errorf("RemoveUserFromGroup: got groupName %q, want %q", groupName, tt.groupName)
						}
						if userID != tt.userID {
							t.Errorf("RemoveUserFromGroup: got userID %q, want %q", userID, tt.userID)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getMemberSchema(t, r)
			raw := makeMemberRaw(
				tftypes.NewValue(tftypes.String, tt.groupName+":"+tt.userID),
				tt.groupName, tt.userID,
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

func TestMemberResource_ImportState(t *testing.T) {
	tests := []struct {
		name        string
		importID    string
		wantGroup   string
		wantUser    string
		expectError bool
	}{
		{
			name:      "valid import ID",
			importID:  "my-group:user-123",
			wantGroup: "my-group",
			wantUser:  "user-123",
		},
		{
			name:        "missing separator",
			importID:    "my-group",
			expectError: true,
		},
		{
			name:        "empty group name",
			importID:    ":user-123",
			expectError: true,
		},
		{
			name:        "empty user ID",
			importID:    "my-group:",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &MemberResource{}
			schResp := getMemberSchema(t, r)

			req := resource.ImportStateRequest{ID: tt.importID}
			resp := &resource.ImportStateResponse{
				State: tfsdk.State{Schema: schResp.Schema, Raw: nullMemberRaw()},
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
			var model MemberResourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.GroupName.ValueString() != tt.wantGroup {
				t.Errorf("GroupName: got %q, want %q", model.GroupName.ValueString(), tt.wantGroup)
			}
			if model.UserID.ValueString() != tt.wantUser {
				t.Errorf("UserID: got %q, want %q", model.UserID.ValueString(), tt.wantUser)
			}
		})
	}
}

func TestMemberResource_Metadata(t *testing.T) {
	r := &MemberResource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "dspc"}, resp)
	if resp.TypeName != "dspc_group_member" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "dspc_group_member")
	}
}

func TestMemberResource_Schema(t *testing.T) {
	r := &MemberResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "group_name", "user_id"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing attribute %q", attr)
		}
	}
}

func TestMemberResource_Configure(t *testing.T) {
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
			r := &MemberResource{}
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

func TestNewMemberResource(t *testing.T) {
	if NewMemberResource() == nil {
		t.Error("NewMemberResource returned nil")
	}
}
