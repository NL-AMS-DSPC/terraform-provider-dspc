package role

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

var dsObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name":    tftypes.String,
		"actions": tftypes.List{ElementType: tftypes.String},
	},
}

func getDataSourceSchema(t *testing.T, d *DataSource) datasource.SchemaResponse {
	t.Helper()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

// makeDSConfigRaw builds a config value as the user would supply it: name is known,
// actions is null because it is computed and not set by the user.
func makeDSConfigRaw(name string) tftypes.Value {
	return tftypes.NewValue(dsObjectType, map[string]tftypes.Value{
		"name":    tftypes.NewValue(tftypes.String, name),
		"actions": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	})
}

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		getRole     func(ctx context.Context, name string) (*client.Role, error)
		expectError bool
	}{
		{
			name:     "successful read with actions",
			roleName: "test-role",
			getRole: func(_ context.Context, name string) (*client.Role, error) {
				return &client.Role{Name: name, Actions: []string{"vm:CreateVM", "uam:ListUsers"}}, nil
			},
		},
		{
			name:     "successful read with no actions",
			roleName: "empty-role",
			getRole: func(_ context.Context, name string) (*client.Role, error) {
				return &client.Role{Name: name, Actions: []string{}}, nil
			},
		},
		{
			name:     "client error",
			roleName: "test-role",
			getRole: func(_ context.Context, _ string) (*client.Role, error) {
				return nil, errors.New("API error 500: internal server error")
			},
			expectError: true,
		},
		{
			name:     "role not found",
			roleName: "missing-role",
			getRole: func(_ context.Context, _ string) (*client.Role, error) {
				return nil, errors.New("resource not found")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d := &DataSource{
				client: &mockResourceClient{
					getRole: tt.getRole,
				},
			}

			schResp := getDataSourceSchema(t, d)
			raw := makeDSConfigRaw(tt.roleName)

			req := datasource.ReadRequest{
				Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw},
			}
			resp := &datasource.ReadResponse{
				State: tfsdk.State{Schema: schResp.Schema},
			}

			d.Read(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Error("expected diagnostics error, got none")
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
			}
			var model DataSourceModel
			if diags := resp.State.Get(ctx, &model); diags.HasError() {
				t.Fatalf("failed to read state: %s", diags)
			}
			if model.Name.ValueString() != tt.roleName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.roleName)
			}
			if model.Actions.IsNull() {
				t.Error("Actions: expected non-null list, got null")
			}
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "asc"}, resp)
	if resp.TypeName != "asc_role" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "asc_role")
	}
}

func TestDataSource_Schema(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"name", "actions"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing attribute %q", attr)
		}
	}
}

func TestDataSource_Configure(t *testing.T) {
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
			d := &DataSource{}
			req := datasource.ConfigureRequest{ProviderData: tt.providerData}
			resp := &datasource.ConfigureResponse{}
			d.Configure(context.Background(), req, resp)
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

func TestNewDataSource(t *testing.T) {
	if NewDataSource() == nil {
		t.Error("NewDataSource returned nil")
	}
}

func TestNewResource(t *testing.T) {
	if NewResource() == nil {
		t.Error("NewResource returned nil")
	}
}
