package objectstorage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/require"
)

var dsObjectTagsType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"key":   tftypes.String,
		"value": tftypes.String,
	},
}

var dsObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":             tftypes.String,
		"name":           tftypes.String,
		"tenant_id":      tftypes.String,
		"reclaim_policy": tftypes.String,
		"endpoint":       tftypes.String,
		"region":         tftypes.String,
		"quota": tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"max_size": tftypes.String,
			},
		},
		"tags": tftypes.List{ElementType: dsObjectTagsType},
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
func makeDSConfigRaw(id string) tftypes.Value {
	quotaValue := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"max_size": tftypes.String,
			},
		},
		map[string]tftypes.Value{
			"max_size": tftypes.NewValue(tftypes.String, "100GB"),
		},
	)
	return tftypes.NewValue(dsObjectType, map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, id),
		"name":           tftypes.NewValue(tftypes.String, nil),
		"tenant_id":      tftypes.NewValue(tftypes.String, nil),
		"reclaim_policy": tftypes.NewValue(tftypes.String, nil),
		"endpoint":       tftypes.NewValue(tftypes.String, nil),
		"region":         tftypes.NewValue(tftypes.String, nil),
		"quota":          quotaValue,
		"tags":           tftypes.NewValue(tftypes.List{ElementType: dsObjectTagsType}, nil),
	})
}

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name            string
		objectStorageID string
		mockResponse    any
		mockStatusCode  int
		objectStorage   DataSourceModel
		expectError     bool
	}{
		{
			name:            "successful read",
			objectStorageID: "some-id",
			mockResponse: &client.ObjectStorage{
				ID:            "some-id",
				Name:          "test-object-storage",
				TenantID:      "test-tenant",
				ReclaimPolicy: "delete",
				Endpoint:      "https://example.com",
				Region:        "us-east-1",
				Quota:         &client.StorageQuota{MaxSize: "100GB"},
				Tags:          []client.Tag{{Key: "group", Value: "test"}},
			},
			mockStatusCode: http.StatusOK,
			objectStorage: DataSourceModel{
				ID:            types.StringValue("some-id"),
				Name:          types.StringValue("test-object-storage"),
				TenantID:      types.StringValue("test-tenant"),
				ReclaimPolicy: types.StringValue("delete"),
				Endpoint:      types.StringValue("https://example.com"),
				Region:        types.StringValue("us-east-1"),
				Quota:         quotaDataModel{MaxSize: types.StringValue("100GB")},
				Tags: []tagModel{
					{
						Key:   types.StringValue("group"),
						Value: types.StringValue("test"),
					},
				},
			},
		},
		{
			name:            "client error",
			objectStorageID: "some-id",
			mockResponse:    map[string]string{"error": "internal server error"},
			mockStatusCode:  http.StatusInternalServerError,
			expectError:     true,
		},
		{
			name:            "object storage not found",
			objectStorageID: "missing-object-storage",
			mockResponse:    map[string]string{"error": "resource not found"},
			mockStatusCode:  http.StatusNotFound,
			expectError:     true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer authServer.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))

			defer server.Close()
			ctx := context.Background()
			d := &DataSource{
				client: client.NewAscClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).ObjectStorage,
			}

			schResp := getDataSourceSchema(t, d)
			raw := makeDSConfigRaw(tt.objectStorageID)

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

			require.Equal(t, tt.objectStorage.ID, model.ID, "ID mismatch")
			require.Equal(t, tt.objectStorage.Name, model.Name, "Name mismatch")
			require.Equal(t, tt.objectStorage.TenantID, model.TenantID, "TenantID mismatch")
			require.Equal(t, tt.objectStorage.ReclaimPolicy, model.ReclaimPolicy, "ReclaimPolicy mismatch")
			require.Equal(t, tt.objectStorage.Endpoint, model.Endpoint, "Endpoint mismatch")
			require.Equal(t, tt.objectStorage.Region, model.Region, "Region mismatch")
			require.Equal(t, tt.objectStorage.Quota.MaxSize, model.Quota.MaxSize, "Quota.MaxSize mismatch")
			require.Equal(t, tt.objectStorage.Tags, model.Tags, "Tags mismatch")
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "asc"}, resp)
	if resp.TypeName != "asc_object_storage" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "asc_object_storage")
	}
}

func TestDataSource_Schema(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}

	// Check for attributes
	for _, attr := range []string{"id", "name", "tenant_id", "reclaim_policy", "endpoint", "region", "tags"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing attribute %q", attr)
		}
	}

	// Check for blocks
	for _, block := range []string{"quota"} {
		if _, ok := resp.Schema.Blocks[block]; !ok {
			t.Errorf("schema missing block %q", block)
		}
	}
}

func TestDataSource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData any
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
