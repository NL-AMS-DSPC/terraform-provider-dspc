package virtualmachine

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVMDataClient implements DataSourceClient and returns a canned list of VMs.
type mockVMDataClient struct {
	response []client.VM
	err      error
}

func (m *mockVMDataClient) ListVMs(_ context.Context) ([]client.VM, error) {
	return m.response, m.err
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	d := &DataSource{}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	t.Run("populates state from client response", func(t *testing.T) {
		mc := &mockVMDataClient{
			response: []client.VM{
				{
					URN:             "vm-urn",
					Name:            "test-vm",
					Status:          "active",
					Tags:            []client.Tag{{Key: "k1", Value: "v1"}},
					AttachedVolumes: []string{"vol-1", "vol-2"},
					SKU: client.SKUResponse{
						ID:          "sku-id",
						Name:        "sku-name",
						Family:      "sku-family",
						Threads:     4,
						Cores:       2,
						MemoryInMB:  1000,
						StorageInGB: 10,
						StorageType: "sku-storage",
						GPUCount:    1,
						GPUType:     "",
					},
					OS: client.OSDetails{
						ID:           "os-id",
						Family:       "os-family",
						Distribution: "os-distribution",
						Release:      "os-release",
						DisplayName:  "os-display-name",
					},
				},
			},
		}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		require.Len(t, out.VirtualMachines, 1)

		vm := out.VirtualMachines[0]
		assert.Equal(t, "test-vm", vm.Name.ValueString())
		assert.Equal(t, "vm-urn", vm.URN.ValueString())
		assert.Equal(t, "active", vm.Status.ValueString())
		assert.Equal(t, "sku-id", vm.SKU.ID.ValueString())
		assert.Equal(t, "sku-name", vm.SKU.Name.ValueString())
		assert.EqualValues(t, 2, vm.SKU.Cores.ValueInt64())
		assert.Equal(t, "os-distribution", vm.OS.Distribution.ValueString())

		var tagsMap map[string]string
		require.False(t, vm.Tags.ElementsAs(ctx, &tagsMap, false).HasError())
		assert.Equal(t, map[string]string{"k1": "v1"}, tagsMap)

		require.Len(t, vm.AttachedVolumes, 2)
		assert.Equal(t, "vol-1", vm.AttachedVolumes[0].ValueString())
		assert.Equal(t, "vol-2", vm.AttachedVolumes[1].ValueString())
	})

	t.Run("empty result produces empty virtual_machines list", func(t *testing.T) {
		mc := &mockVMDataClient{response: []client.VM{}}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		assert.Empty(t, out.VirtualMachines)
	})

	t.Run("client error becomes diagnostic error", func(t *testing.T) {
		mc := &mockVMDataClient{err: assert.AnError}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		assert.True(t, resp.Diagnostics.HasError())
	})
}

func TestMetadata(t *testing.T) {
	dataSource := &DataSource{}

	req := datasource.MetadataRequest{ProviderTypeName: "asc"}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)
	assert.Equal(t, "asc_virtual_machines", resp.TypeName)
}

func TestSchema(t *testing.T) {
	dataSource := &DataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Data source schema has errors: %v", resp.Diagnostics)
	}

	if resp.Schema.Attributes == nil {
		t.Error("Data source schema attributes is nil")
	}

	// Check that virtual_machines attribute exists
	attributes := resp.Schema.Attributes
	if _, ok := attributes["virtual_machines"]; !ok {
		t.Error("Data source schema missing 'virtual_machines' attribute")
	}
}

func TestConfigure(t *testing.T) {
	tests := []struct {
		name         string
		providerData any
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: client.NewDspcClient("http://localhost", "test-ns", "test-client-id", "test-client-secret", "http://auth.localhost", "test-realm", 30),
			expectError:  false,
		},
		{
			name:         "nil provider data",
			providerData: nil,
			expectError:  false, // Should not error, just skip configuration
		},
		{
			name:         "invalid provider data type",
			providerData: "not-a-client",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSource := &DataSource{}

			req := datasource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			dataSource.Configure(context.Background(), req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error, got none")
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no error, got: %v", resp.Diagnostics)
				}
			}
		})
	}
}
