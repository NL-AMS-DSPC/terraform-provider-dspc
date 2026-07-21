package vmgroup

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVMGroupDataClient implements DataClient and returns a canned list of VM groups.
type mockVMGroupDataClient struct {
	response []client.VMGroup
	err      error
}

func (m *mockVMGroupDataClient) ListVMGroups(_ context.Context) ([]client.VMGroup, error) {
	return m.response, m.err
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	d := &DataSource{}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	t.Run("populates state from client response", func(t *testing.T) {
		mc := &mockVMGroupDataClient{
			response: []client.VMGroup{
				{
					URN:             "vm-group-urn",
					Name:            "test-vm-group",
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
					AutoscalingPolicy: &client.AutoscalingPolicy{
						MinReplicas: 1,
						MaxReplicas: 5,
						CronRule: &client.CronRule{
							Timezone:        "UTC",
							Start:           "0 8 * * *",
							End:             "0 18 * * *",
							DesiredReplicas: 3,
						},
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
		require.Len(t, out.VMGroups, 1)

		vmGroup := out.VMGroups[0]
		assert.Equal(t, "test-vm-group", vmGroup.Name.ValueString())
		assert.Equal(t, "vm-group-urn", vmGroup.URN.ValueString())
		assert.Equal(t, "active", vmGroup.Status.ValueString())
		assert.Equal(t, "sku-id", vmGroup.SKU.ID.ValueString())
		assert.Equal(t, "sku-name", vmGroup.SKU.Name.ValueString())
		assert.Equal(t, int64(2), vmGroup.SKU.Cores.ValueInt64())

		var tagsMap map[string]string
		require.False(t, vmGroup.Tags.ElementsAs(ctx, &tagsMap, false).HasError())
		assert.Equal(t, map[string]string{"k1": "v1"}, tagsMap)

		require.Len(t, vmGroup.AttachedVolumes, 2)
		assert.Equal(t, "vol-1", vmGroup.AttachedVolumes[0].ValueString())
		assert.Equal(t, "vol-2", vmGroup.AttachedVolumes[1].ValueString())

		require.NotNil(t, vmGroup.AutoscalingPolicy)
		assert.Equal(t, int32(1), vmGroup.AutoscalingPolicy.MinReplicas.ValueInt32())
		assert.Equal(t, int32(5), vmGroup.AutoscalingPolicy.MaxReplicas.ValueInt32())

		require.NotNil(t, vmGroup.AutoscalingPolicy.CronRule)
		assert.Equal(t, "UTC", vmGroup.AutoscalingPolicy.CronRule.Timezone.ValueString())
		assert.Equal(t, "0 8 * * *", vmGroup.AutoscalingPolicy.CronRule.Start.ValueString())
		assert.Equal(t, "0 18 * * *", vmGroup.AutoscalingPolicy.CronRule.End.ValueString())
		assert.Equal(t, int32(3), vmGroup.AutoscalingPolicy.CronRule.DesiredReplicas.ValueInt32())
	})

	t.Run("nil autoscaling policy produces nil model", func(t *testing.T) {
		mc := &mockVMGroupDataClient{
			response: []client.VMGroup{
				{
					URN:  "vm-group-urn",
					Name: "test-vm-group",
				},
			},
		}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		require.Len(t, out.VMGroups, 1)
		assert.Nil(t, out.VMGroups[0].AutoscalingPolicy)
	})

	t.Run("empty result produces empty vm_groups list", func(t *testing.T) {
		mc := &mockVMGroupDataClient{response: []client.VMGroup{}}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		assert.Empty(t, out.VMGroups)
	})

	t.Run("client error becomes diagnostic error", func(t *testing.T) {
		mc := &mockVMGroupDataClient{err: assert.AnError}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		assert.True(t, resp.Diagnostics.HasError())
	})
}

func TestMetadata(t *testing.T) {
	dataSource := &DataSource{}

	req := datasource.MetadataRequest{ProviderTypeName: "dspc"}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)
	assert.Equal(t, "dspc_vm_groups", resp.TypeName)
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

	// Check that vm_groups attribute exists
	attributes := resp.Schema.Attributes
	if _, ok := attributes["vm_groups"]; !ok {
		t.Error("Data source schema missing 'vm_groups' attribute")
	}
}

func TestConfigure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
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
