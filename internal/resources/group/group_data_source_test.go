package group

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
)

var groupDSObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String,
	},
}

func getGroupDSSchema(t *testing.T, d *DataSource) datasource.SchemaResponse {
	t.Helper()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp
}

func makeGroupDSConfigRaw(name string) tftypes.Value {
	return tftypes.NewValue(groupDSObjectType, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, name),
	})
}

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name        string
		groupName   string
		getGroup    func(ctx context.Context, name string) (*client.Group, error)
		expectError bool
	}{
		{
			name:      "successful read",
			groupName: "test-group",
			getGroup: func(_ context.Context, name string) (*client.Group, error) {
				return &client.Group{Name: name}, nil
			},
		},
		{
			name:      "client error",
			groupName: "test-group",
			getGroup: func(_ context.Context, _ string) (*client.Group, error) {
				return nil, errors.New("API error 500: internal server error")
			},
			expectError: true,
		},
		{
			name:      "group not found",
			groupName: "missing-group",
			getGroup: func(_ context.Context, _ string) (*client.Group, error) {
				return nil, errors.New("resource not found")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d := &DataSource{
				client: &mockGroupResourceClient{
					getGroup: tt.getGroup,
				},
			}

			schResp := getGroupDSSchema(t, d)
			raw := makeGroupDSConfigRaw(tt.groupName)

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
			if model.Name.ValueString() != tt.groupName {
				t.Errorf("Name: got %q, want %q", model.Name.ValueString(), tt.groupName)
			}
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "asc"}, resp)
	if resp.TypeName != "asc_group" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "asc_group")
	}
}

func TestDataSource_Schema(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["name"]; !ok {
		t.Error("schema missing attribute \"name\"")
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
