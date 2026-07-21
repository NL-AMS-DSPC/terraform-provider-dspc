package container

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

const (
	containerPath = "/api/containers/v1/deployments"
)

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		container      client.Container
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful creation",
			container: client.Container{
				Name: "test-container",
			},
			mockResponse: map[string]any{"data": &client.Container{
				Name: "test-container",
			}},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "API error - conflict",
			container: client.Container{
				Name: "existing-container",
			},
			mockResponse:   map[string]string{"error": "Container name already exists"},
			mockStatusCode: http.StatusConflict,
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

			containerResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}

			container, err := containerResource.client.CreateDeployment(
				context.Background(),
				tt.container,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.container.Name, container.Name)
			}
		})
	}
}

func TestResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		containerName  string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			containerName:  "test-container",
			mockResponse:   map[string]string{"deleted": "test-container"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error - not found",
			containerName:  "nonexistent-container",
			mockResponse:   map[string]string{"error": "not found"},
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

			containerResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}
			err := containerResource.client.DeleteDeployment(context.Background(), tt.containerName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResource_ImportState(t *testing.T) {
	tests := []struct {
		name           string
		importID       string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "successful import",
			importID: "test-container",
			mockResponse: map[string]any{"data": &client.Container{
				Name: "test-container",
			}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "import non-existent container",
			importID:       "nonexistent-container",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "API error during import",
			importID:       "test-container",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != containerPath+"/"+tt.importID {
					t.Fatalf("Expected %s path, got %s", containerPath+"/"+tt.importID, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			containerResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Containers,
			}
			container, err := containerResource.client.GetDeployment(context.Background(), tt.importID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.importID, container.Name)
			}
		})
	}
}

// mockResourceClient implements ResourceClient with function fields for test control.
type mockResourceClient struct {
	createDeployment func(ctx context.Context, req client.Container) (*client.Container, error)
	getDeployment    func(ctx context.Context, name string) (*client.Container, error)
	patchDeployment  func(ctx context.Context, name string, req client.PatchTagsRequest) (*client.Container, error)
	deleteDeployment func(ctx context.Context, name string) error
}

func (m *mockResourceClient) CreateDeployment(ctx context.Context, req client.Container) (*client.Container, error) {
	return m.createDeployment(ctx, req)
}

func (m *mockResourceClient) GetDeployment(ctx context.Context, name string) (*client.Container, error) {
	return m.getDeployment(ctx, name)
}

func (m *mockResourceClient) PatchDeployment(ctx context.Context, name string, req client.PatchTagsRequest) (*client.Container, error) {
	return m.patchDeployment(ctx, name, req)
}

func (m *mockResourceClient) DeleteDeployment(ctx context.Context, name string) error {
	return m.deleteDeployment(ctx, name)
}

var (
	containerRegistryAuthType = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"server":   tftypes.String,
			"username": tftypes.String,
			"password": tftypes.String,
		},
	}
	containerSecretType = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"env_name": tftypes.String,
			"value":    tftypes.String,
		},
	}
	containerObjectType = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":            tftypes.String,
			"name":          tftypes.String,
			"tenant_id":     tftypes.String,
			"image":         tftypes.String,
			"sku_id":        tftypes.String,
			"port":          tftypes.Number,
			"command":       tftypes.String,
			"args":          tftypes.List{ElementType: tftypes.String},
			"env":           tftypes.List{ElementType: tftypes.String},
			"working_dir":   tftypes.String,
			"user":          tftypes.String,
			"group":         tftypes.String,
			"replicas":      tftypes.Number,
			"tags":          tftypes.Map{ElementType: tftypes.String},
			"registry_auth": containerRegistryAuthType,
			"secrets":       tftypes.List{ElementType: containerSecretType},
		},
	}
)

func getResourceSchema(t *testing.T, r *Resource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

// makeContainerRaw builds a fully-typed container object value. A nil tags map yields
// a null tags attribute; a nil args/env slice yields a null list. Unused optional
// attributes are null.
func makeContainerRaw(name string, tags map[string]string, args, env []string) tftypes.Value {
	stringListType := tftypes.List{ElementType: tftypes.String}

	tagsVal := tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil)
	if tags != nil {
		m := make(map[string]tftypes.Value, len(tags))
		for k, v := range tags {
			m[k] = tftypes.NewValue(tftypes.String, v)
		}
		tagsVal = tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, m)
	}

	stringList := func(items []string) tftypes.Value {
		if items == nil {
			return tftypes.NewValue(stringListType, nil)
		}
		vals := make([]tftypes.Value, len(items))
		for i, it := range items {
			vals[i] = tftypes.NewValue(tftypes.String, it)
		}
		return tftypes.NewValue(stringListType, vals)
	}

	return tftypes.NewValue(containerObjectType, map[string]tftypes.Value{
		"id":            tftypes.NewValue(tftypes.String, name),
		"name":          tftypes.NewValue(tftypes.String, name),
		"tenant_id":     tftypes.NewValue(tftypes.String, "tenant-1"),
		"image":         tftypes.NewValue(tftypes.String, "nginx:latest"),
		"sku_id":        tftypes.NewValue(tftypes.String, "gp-1"),
		"port":          tftypes.NewValue(tftypes.Number, int64(8080)),
		"command":       tftypes.NewValue(tftypes.String, nil),
		"args":          stringList(args),
		"env":           stringList(env),
		"working_dir":   tftypes.NewValue(tftypes.String, nil),
		"user":          tftypes.NewValue(tftypes.String, nil),
		"group":         tftypes.NewValue(tftypes.String, nil),
		"replicas":      tftypes.NewValue(tftypes.Number, int64(1)),
		"tags":          tagsVal,
		"registry_auth": tftypes.NewValue(containerRegistryAuthType, nil),
		"secrets":       tftypes.NewValue(tftypes.List{ElementType: containerSecretType}, nil),
	})
}

func TestComputeTagPatch(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name  string
		state map[string]string
		plan  map[string]string
		want  []client.PatchTagDTO
	}{
		{
			name:  "add only",
			state: map[string]string{"a": "1"},
			plan:  map[string]string{"a": "1", "b": "2"},
			want:  []client.PatchTagDTO{{Key: "b", Value: strPtr("2")}},
		},
		{
			name:  "change only",
			state: map[string]string{"a": "1"},
			plan:  map[string]string{"a": "2"},
			want:  []client.PatchTagDTO{{Key: "a", Value: strPtr("2")}},
		},
		{
			name:  "delete only",
			state: map[string]string{"a": "1", "b": "2"},
			plan:  map[string]string{"a": "1"},
			want:  []client.PatchTagDTO{{Key: "b", Value: nil}},
		},
		{
			name:  "mixed add change delete (upserts sorted, then deletions sorted)",
			state: map[string]string{"keep": "same", "change": "old", "drop2": "x", "drop1": "y"},
			plan:  map[string]string{"keep": "same", "change": "new", "add": "z"},
			want: []client.PatchTagDTO{
				{Key: "add", Value: strPtr("z")},
				{Key: "change", Value: strPtr("new")},
				{Key: "drop1", Value: nil},
				{Key: "drop2", Value: nil},
			},
		},
		{
			name:  "no change",
			state: map[string]string{"a": "1"},
			plan:  map[string]string{"a": "1"},
			want:  []client.PatchTagDTO{},
		},
		{
			name:  "state nil plan set",
			state: nil,
			plan:  map[string]string{"a": "1"},
			want:  []client.PatchTagDTO{{Key: "a", Value: strPtr("1")}},
		},
		{
			name:  "plan nil state set (all deletions)",
			state: map[string]string{"a": "1", "b": "2"},
			plan:  nil,
			want: []client.PatchTagDTO{
				{Key: "a", Value: nil},
				{Key: "b", Value: nil},
			},
		},
		{
			name:  "both nil",
			state: nil,
			plan:  nil,
			want:  []client.PatchTagDTO{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeTagPatch(tt.state, tt.plan)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeTagPatch() = %s, want %s", formatPatch(got), formatPatch(tt.want))
			}
		})
	}
}

// formatPatch renders a patch slice with dereferenced values for readable failures.
func formatPatch(patch []client.PatchTagDTO) string {
	out := "["
	for i, p := range patch {
		if i > 0 {
			out += ", "
		}
		if p.Value == nil {
			out += p.Key + "=<nil>"
		} else {
			out += p.Key + "=" + *p.Value
		}
	}
	return out + "]"
}

func TestResource_Update(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		stateTags   map[string]string // nil means the tags attribute is null
		planTags    map[string]string
		stateArgs   []string
		planArgs    []string
		patchErr    error
		expectCall  bool
		expectBody  []client.PatchTagDTO
		expectError bool
	}{
		{
			name:       "mixed add change delete",
			stateTags:  map[string]string{"keep": "same", "change": "old", "drop": "x"},
			planTags:   map[string]string{"keep": "same", "change": "new", "add": "z"},
			expectCall: true,
			expectBody: []client.PatchTagDTO{
				{Key: "add", Value: strPtr("z")},
				{Key: "change", Value: strPtr("new")},
				{Key: "drop", Value: nil},
			},
		},
		{
			name:       "all tags removed",
			stateTags:  map[string]string{"a": "1", "b": "2"},
			planTags:   nil,
			expectCall: true,
			expectBody: []client.PatchTagDTO{
				{Key: "a", Value: nil},
				{Key: "b", Value: nil},
			},
		},
		{
			name:       "no tag change",
			stateTags:  map[string]string{"a": "1"},
			planTags:   map[string]string{"a": "1"},
			expectCall: false,
		},
		{
			name:        "args changed is rejected",
			stateTags:   map[string]string{"a": "1"},
			planTags:    map[string]string{"a": "1", "b": "2"},
			stateArgs:   []string{"--old"},
			planArgs:    []string{"--new"},
			expectCall:  false,
			expectError: true,
		},
		{
			name:        "api error",
			stateTags:   map[string]string{"a": "1"},
			planTags:    map[string]string{"a": "2"},
			patchErr:    errors.New("API error 500: internal server error"),
			expectCall:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			var calls int
			var gotBody []client.PatchTagDTO
			r := &Resource{
				client: &mockResourceClient{
					patchDeployment: func(_ context.Context, name string, req client.PatchTagsRequest) (*client.Container, error) {
						calls++
						gotBody = req.Tags
						if tt.patchErr != nil {
							return nil, tt.patchErr
						}
						// Echo the plan's resulting tags back as the fresh deployment.
						resp := &client.Container{Name: name}
						for _, m := range req.Tags {
							if m.Value != nil {
								resp.Tags = append(resp.Tags, client.ContainerTag{Key: m.Key, Value: *m.Value})
							}
						}
						return resp, nil
					},
				},
			}

			schResp := getResourceSchema(t, r)
			planRaw := makeContainerRaw("test-container", tt.planTags, tt.planArgs, nil)
			stateRaw := makeContainerRaw("test-container", tt.stateTags, tt.stateArgs, nil)

			req := resource.UpdateRequest{
				Plan:  tfsdk.Plan{Schema: schResp.Schema, Raw: planRaw},
				State: tfsdk.State{Schema: schResp.Schema, Raw: stateRaw},
			}
			resp := &resource.UpdateResponse{
				State: tfsdk.State{Schema: schResp.Schema, Raw: stateRaw},
			}

			r.Update(ctx, req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Fatal("expected diagnostics error, got none")
				}
			} else if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics error: %s", resp.Diagnostics)
			}

			if tt.expectCall {
				if calls != 1 {
					t.Errorf("expected exactly 1 PatchDeployment call, got %d", calls)
				}
				if tt.expectBody != nil && !reflect.DeepEqual(gotBody, tt.expectBody) {
					t.Errorf("patch body = %s, want %s", formatPatch(gotBody), formatPatch(tt.expectBody))
				}
			} else if calls != 0 {
				t.Errorf("expected zero PatchDeployment calls, got %d", calls)
			}

			// On the successful, non-rejected paths, final state tags must equal the plan.
			if !tt.expectError {
				var model ResourceModel
				if diags := resp.State.Get(ctx, &model); diags.HasError() {
					t.Fatalf("failed to read state: %s", diags)
				}
				var gotTags map[string]string
				if !model.Tags.IsNull() {
					if diags := model.Tags.ElementsAs(ctx, &gotTags, false); diags.HasError() {
						t.Fatalf("failed to read tags: %s", diags)
					}
				}
				if !reflect.DeepEqual(gotTags, tt.planTags) {
					t.Errorf("state tags = %v, want %v", gotTags, tt.planTags)
				}
			}
		})
	}
}

func TestMapStateFromContainer_NoTags(t *testing.T) {
	ctx := context.Background()
	var model ResourceModel
	var diags diag.Diagnostics

	mapStateFromContainer(ctx, &model, &client.Container{
		Name:  "test-container",
		Image: "nginx:latest",
		Port:  8080,
	}, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics error: %s", diags)
	}
	if !model.Tags.IsNull() {
		t.Errorf("expected Tags to be null when the API returns zero tags, got %v", model.Tags)
	}
}
