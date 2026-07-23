package sku

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSKUDataClient implements DataClient and returns a canned list of SKUs.
type mockSKUDataClient struct {
	response []client.SKUResponse
	err      error
}

func (m *mockSKUDataClient) ListSKUs(_ context.Context) ([]client.SKUResponse, error) {
	return m.response, m.err
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	d := &DataSource{}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	t.Run("populates state from client response", func(t *testing.T) {
		d.client = &mockSKUDataClient{
			response: []client.SKUResponse{
				{
					ID:          "sku-id",
					Name:        "sku-name",
					Family:      "sku-family",
					Threads:     4,
					Cores:       2,
					MemoryInMB:  1000,
					StorageInGB: 10,
					StorageType: "sku-storage",
					GPUCount:    1,
				},
			},
		}

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		require.Len(t, out.SKUs, 1)

		s := out.SKUs[0]
		assert.Equal(t, "sku-id", s.ID.ValueString())
		assert.Equal(t, "sku-name", s.Name.ValueString())
		assert.Equal(t, "sku-family", s.Family.ValueString())
		assert.EqualValues(t, 4, s.Threads.ValueInt64())
		assert.EqualValues(t, 2, s.Cores.ValueInt64())
		assert.EqualValues(t, 1000, s.MemoryInMB.ValueInt64())
		assert.EqualValues(t, 10, s.StorageInGB.ValueInt64())
		assert.Equal(t, "sku-storage", s.StorageType.ValueString())
		assert.EqualValues(t, 1, s.GPUCount.ValueInt64())
	})

	t.Run("empty result produces empty skus list", func(t *testing.T) {
		d.client = &mockSKUDataClient{response: []client.SKUResponse{}}

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		assert.Empty(t, out.SKUs)
	})

	t.Run("client error becomes diagnostic error", func(t *testing.T) {
		d.client = &mockSKUDataClient{err: assert.AnError}

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
	assert.Equal(t, "asc_skus", resp.TypeName)
}

func TestSchema(t *testing.T) {
	dataSource := &DataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.Schema.Attributes)

	_, ok := resp.Schema.Attributes["skus"]
	assert.True(t, ok, "Data source schema missing 'skus' attribute")
}

func TestConfigure(t *testing.T) {
	tests := []struct {
		name         string
		providerData any
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: client.NewAscClient("http://localhost", "test-ns", "test-client-id", "test-client-secret", "http://auth.localhost", "test-realm", 30),
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
				assert.True(t, resp.Diagnostics.HasError())
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}
