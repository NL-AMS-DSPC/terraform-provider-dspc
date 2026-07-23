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

// mockGroupResourceClient implements ResourceClient with function fields for test control.
type mockGroupResourceClient struct {
	createGroup func(ctx context.Context, name string) error
	getGroup    func(ctx context.Context, name string) (*client.Group, error)
	deleteGroup func(ctx context.Context, name string) error
}

func (m *mockGroupResourceClient) CreateGroup(ctx context.Context, name string) error {
	return m.createGroup(ctx, name)
}

func (m *mockGroupResourceClient) GetGroup(ctx context.Context, name string) (*client.Group, error) {
	return m.getGroup(ctx, name)
}

func (m *mockGroupResourceClient) DeleteGroup(ctx context.Context, name string) error {
	return m.deleteGroup(ctx, name)
}

var groupObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	},
}

func getGroupSchema(t *testing.T, r *Resource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func makeGroupRaw(id tftypes.Value, name string) tftypes.Value {
	return tftypes.NewValue(groupObjectType, map[string]tftypes.Value{
		"id":   id,
		"name": tftypes.NewValue(tftypes.String, name),
	})
}

func nullGroupRaw() tftypes.Value {
	return tftypes.NewValue(groupObjectType, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, nil),
		"name": tftypes.NewValue(tftypes.String, nil),
	})
}

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		clientErr   error
		expectError bool
	}{
		{
			name:      "successful creation",
			groupName: "test-group",
		},
		{
			name:        "client error",
			groupName:   "bad-group",
			clientErr:   errors.New("API error 409: group already exists"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockGroupResourceClient{
					createGroup: func(_ context.Context, name string) error {
						if name != tt.groupName {
							t.Errorf("CreateGroup: got name %q, want %q", name, tt.groupName)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getGroupSchema(t, r)
			raw := makeGroupRaw(tftypes.NewValue(tftypes.String, tftypes.UnknownValue), tt.groupName)

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
			if model.ID.ValueString() != tt.groupName {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), tt.groupName)
			}
			if model.Name.ValueString() != tt.groupName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.groupName)
			}
		})
	}
}

func TestResource_Read(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		getGroup    func(ctx context.Context, name string) (*client.Group, error)
		expectError bool
		expectGone  bool
	}{
		{
			name:      "successful read",
			groupName: "test-group",
			getGroup: func(_ context.Context, name string) (*client.Group, error) {
				return &client.Group{Name: name}, nil
			},
		},
		{
			name:      "group not found removes from state",
			groupName: "gone-group",
			getGroup: func(_ context.Context, _ string) (*client.Group, error) {
				return nil, errors.New("resource not found")
			},
			expectGone: true,
		},
		{
			name:      "API 404 removes from state",
			groupName: "gone-group",
			getGroup: func(_ context.Context, _ string) (*client.Group, error) {
				return nil, errors.New("API error 404: not found")
			},
			expectGone: true,
		},
		{
			name:      "client error",
			groupName: "test-group",
			getGroup: func(_ context.Context, _ string) (*client.Group, error) {
				return nil, errors.New("connection refused")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockGroupResourceClient{
					getGroup: tt.getGroup,
				},
			}

			schResp := getGroupSchema(t, r)
			raw := makeGroupRaw(tftypes.NewValue(tftypes.String, tt.groupName), tt.groupName)

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
			if model.ID.ValueString() != tt.groupName {
				t.Errorf("ID: got %q, want %q", model.ID.ValueString(), tt.groupName)
			}
			if model.Name.ValueString() != tt.groupName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.groupName)
			}
		})
	}
}

func TestResource_Update(t *testing.T) {
	r := &Resource{}
	resp := &resource.UpdateResponse{}
	r.Update(context.Background(), resource.UpdateRequest{}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected Update to always return a diagnostics error, got none")
	}
}

func TestResource_Delete(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		clientErr   error
		expectError bool
	}{
		{
			name:      "successful deletion",
			groupName: "test-group",
		},
		{
			name:      "group not found — silently removed from state",
			groupName: "gone-group",
			clientErr: errors.New("resource not found"),
		},
		{
			name:        "API error",
			groupName:   "test-group",
			clientErr:   errors.New("API error 500: internal server error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			r := &Resource{
				client: &mockGroupResourceClient{
					deleteGroup: func(_ context.Context, name string) error {
						if name != tt.groupName {
							t.Errorf("DeleteGroup: got name %q, want %q", name, tt.groupName)
						}
						return tt.clientErr
					},
				},
			}

			schResp := getGroupSchema(t, r)
			raw := makeGroupRaw(tftypes.NewValue(tftypes.String, tt.groupName), tt.groupName)

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
	schResp := getGroupSchema(t, r)

	req := resource.ImportStateRequest{ID: "my-group"}
	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: schResp.Schema, Raw: nullGroupRaw()},
	}

	r.ImportState(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
	}
	var model ResourceModel
	if diags := resp.State.Get(ctx, &model); diags.HasError() {
		t.Fatalf("failed to read state: %s", diags)
	}
	if model.Name.ValueString() != "my-group" {
		t.Errorf("Name: got %q, want %q", model.Name.ValueString(), "my-group")
	}
}

func TestResource_Metadata(t *testing.T) {
	r := &Resource{}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "asc"}, resp)
	if resp.TypeName != "asc_group" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "asc_group")
	}
}

func TestResource_Schema(t *testing.T) {
	r := &Resource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "name"} {
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

func TestNewResource(t *testing.T) {
	if NewResource() == nil {
		t.Error("NewResource returned nil")
	}
}
