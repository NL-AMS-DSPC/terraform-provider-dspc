package container

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

var dsObjectTagsType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"key":   tftypes.String,
		"value": tftypes.String,
	},
}

var dsObjectType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":          tftypes.String,
		"name":        tftypes.String,
		"image":       tftypes.String,
		"port":        tftypes.Number,
		"command":     tftypes.String,
		"args":        tftypes.List{ElementType: tftypes.String},
		"env":         tftypes.List{ElementType: tftypes.String},
		"working_dir": tftypes.String,
		"user":        tftypes.String,
		"group":       tftypes.String,
		"replicas":    tftypes.Number,
		"tags":        tftypes.List{ElementType: dsObjectTagsType},
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
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, name),
		"image":       tftypes.NewValue(tftypes.String, nil),
		"port":        tftypes.NewValue(tftypes.Number, nil),
		"command":     tftypes.NewValue(tftypes.String, nil),
		"args":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"env":         tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"working_dir": tftypes.NewValue(tftypes.String, nil),
		"user":        tftypes.NewValue(tftypes.String, nil),
		"group":       tftypes.NewValue(tftypes.String, nil),
		"replicas":    tftypes.NewValue(tftypes.Number, nil),
		"tags":        tftypes.NewValue(tftypes.List{ElementType: dsObjectTagsType}, nil),
	})
}

func TestDataSource_Read(t *testing.T) {
	tagsType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}}
	tests := []struct {
		name           string
		containerName  string
		mockResponse   any
		mockStatusCode int
		container      DataSourceModel
		expectError    bool
	}{
		{
			name:          "successful read",
			containerName: "test-container",
			mockResponse: map[string]any{"data": &client.Container{
				ID:         "some-id",
				Name:       "test-container",
				Image:      "sample-image",
				Port:       1234,
				Command:    "cowsay",
				Args:       []string{"hello world"},
				Env:        []string{"SOME_VAR=1"},
				WorkingDir: "/home/john",
				User:       "john",
				Group:      "users",
				Replicas:   1,
				Tags:       []client.ContainerTag{{Key: "group", Value: "test"}},
			}},
			mockStatusCode: http.StatusOK,
			container: DataSourceModel{
				ID:         types.StringValue("some-id"),
				Name:       types.StringValue("test-container"),
				Image:      types.StringValue("sample-image"),
				Port:       types.Int32Value(1234),
				Command:    types.StringValue("cowsay"),
				Args:       types.ListValueMust(types.StringType, []attr.Value{types.StringValue("hello world")}),
				Env:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("SOME_VAR=1")}),
				WorkingDir: types.StringValue("/home/john"),
				User:       types.StringValue("john"),
				Group:      types.StringValue("users"),
				Replicas:   types.Int32Value(1),
				Tags:       types.ListValueMust(tagsType, []attr.Value{types.ObjectValueMust(tagsType.AttrTypes, map[string]attr.Value{"key": types.StringValue("group"), "value": types.StringValue("test")})}),
			},
		},
		{
			name:           "client error",
			containerName:  "test-container",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "container not found",
			containerName:  "missing-container",
			mockResponse:   map[string]string{"error": "resource not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
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
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}

			schResp := getDataSourceSchema(t, d)
			raw := makeDSConfigRaw(tt.containerName)

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

			if model.ID != tt.container.ID {
				t.Errorf("ID: got %q, want %q", model.ID, tt.container.ID)
			}
			if model.Name != tt.container.Name {
				t.Errorf("Name: got %q, want %q", model.Name, tt.container.Name)
			}
			if model.Image != tt.container.Image {
				t.Errorf("Image: got %q, want %q", model.Image, tt.container.Image)
			}
			if model.Port != tt.container.Port {
				t.Errorf("Port: got %q, want %q", model.Port, tt.container.Port)
			}
			if model.Command != tt.container.Command {
				t.Errorf("Command: got %q, want %q", model.Command, tt.container.Command)
			}
			if !reflect.DeepEqual(model.Args, tt.container.Args) {
				t.Errorf("Args: got %q, want %q", model.Args, tt.container.Args)
			}
			if !reflect.DeepEqual(model.Env, tt.container.Env) {
				t.Errorf("Env: got %q, want %q", model.Env, tt.container.Env)
			}
			if model.WorkingDir != tt.container.WorkingDir {
				t.Errorf("WorkingDir: got %q, want %q", model.WorkingDir, tt.container.WorkingDir)
			}
			if model.User != tt.container.User {
				t.Errorf("User: got %q, want %q", model.User, tt.container.User)
			}
			if model.Group != tt.container.Group {
				t.Errorf("Group: got %q, want %q", model.Group, tt.container.Group)
			}
			if model.Replicas != tt.container.Replicas {
				t.Errorf("Replicas: got %q, want %q", model.Replicas, tt.container.Replicas)
			}
			if !reflect.DeepEqual(model.Tags, tt.container.Tags) {
				t.Errorf("Tags: got %q, want %q", model.Tags, tt.container.Tags)
			}
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "dspc"}, resp)
	if resp.TypeName != "dspc_container" {
		t.Errorf("TypeName: got %q, want %q", resp.TypeName, "dspc_container")
	}
}

func TestDataSource_Schema(t *testing.T) {
	d := &DataSource{}
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), datasource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected schema error: %s", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "name", "image", "port", "command", "args", "env", "working_dir", "user", "group", "replicas", "tags"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("schema missing attribute %q", attr)
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
