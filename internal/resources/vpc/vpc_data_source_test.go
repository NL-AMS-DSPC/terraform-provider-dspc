package vpc

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/subnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDataClient implements DataClient and returns a canned list of VPCs.
type mockDataClient struct {
	response []client.VPC
	err      error
}

func (m *mockDataClient) ListVPCs(_ context.Context) ([]client.VPC, error) {
	return m.response, m.err
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	d := &DataSource{}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	t.Run("populates state from client response", func(t *testing.T) {
		mc := &mockDataClient{
			response: []client.VPC{
				{
					ID:     "vpc-id",
					URN:    "vpc-urn",
					Name:   "test-vpc",
					CIDR:   "10.0.0.0/24",
					Status: "active",
					Tags:   []client.Tag{{Key: "k1", Value: "v1"}},
					Subnets: []client.Subnet{
						{
							ID:     "subnet-id",
							URN:    "subnet-urn",
							Name:   "test-subnet",
							CIDR:   "10.0.0.0/25",
							Type:   "public",
							VPCID:  "vpc-id",
							Status: "active",
							Tags:   []client.Tag{{Key: "sk1", Value: "sv1"}},
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
		require.Len(t, out.VPCs, 1)
		assert.Equal(t, "test-vpc", out.VPCs[0].Name.ValueString())
		assert.Equal(t, "10.0.0.0/24", out.VPCs[0].CIDR.ValueString())
		assert.Equal(t, "active", out.VPCs[0].Status.ValueString())

		var tagsMap map[string]string
		require.False(t, out.VPCs[0].Tags.ElementsAs(ctx, &tagsMap, false).HasError())
		assert.Equal(t, map[string]string{"k1": "v1"}, tagsMap)

		var subnets []subnet.Model
		require.False(t, out.VPCs[0].Subnets.ElementsAs(ctx, &subnets, false).HasError())
		require.Len(t, subnets, 1)
		assert.Equal(t, "test-subnet", subnets[0].Name.ValueString())
		assert.Equal(t, "subnet-urn", subnets[0].URN.ValueString())

		require.False(t, subnets[0].Tags.ElementsAs(ctx, &tagsMap, false).HasError())
		assert.Equal(t, map[string]string{"sk1": "sv1"}, tagsMap)
	})

	t.Run("empty result produces empty vpcs list", func(t *testing.T) {
		mc := &mockDataClient{response: []client.VPC{}}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		assert.Empty(t, out.VPCs)
	})

	t.Run("client error becomes diagnostic error", func(t *testing.T) {
		mc := &mockDataClient{err: assert.AnError}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{}, resp)
		assert.True(t, resp.Diagnostics.HasError())
	})
}
