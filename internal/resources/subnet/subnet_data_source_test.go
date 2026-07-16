package subnet

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDataClient implements DataClient and records the vpc_name it was called with.
type mockDataClient struct {
	gotVPCName string
	response   []client.Subnet
	err        error
}

func (m *mockDataClient) ListSubnetsForVPC(_ context.Context, vpcName string) ([]client.Subnet, error) {
	m.gotVPCName = vpcName
	return m.response, m.err
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	d := &DataSource{}

	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	listType, ok := schemaResp.Schema.Attributes["subnets"].GetType().(types.ListType)
	if !ok {
		t.Errorf("failed to get ListType from schema for subnets")
	}
	subnetsElemType := listType.ElemType

	buildConfig := func(vpcName string) tfsdk.Config {
		plan := tfsdk.Plan{Schema: schemaResp.Schema}
		diags := plan.Set(ctx, &DataSourceModel{
			VPCName: types.StringValue(vpcName),
			Subnets: types.ListNull(subnetsElemType),
		})
		require.False(t, diags.HasError(), diags)
		return tfsdk.Config{Schema: schemaResp.Schema, Raw: plan.Raw}
	}

	t.Run("populates state from client response", func(t *testing.T) {
		mc := &mockDataClient{
			response: []client.Subnet{
				{
					ID:     "subnet-id",
					URN:    "subnet-urn",
					Name:   "test-subnet",
					CIDR:   "10.0.0.0/25",
					Type:   "public",
					VPCID:  "test-vpc-id",
					Status: "active",
					Tags:   []client.Tag{{Key: "k1", Value: "v1"}},
				},
			},
		}
		d.client = mc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{Config: buildConfig("test-vpc")}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		assert.Equal(t, "test-vpc", mc.gotVPCName)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())

		var models []Model
		require.False(t, out.Subnets.ElementsAs(ctx, &models, false).HasError())
		require.Len(t, models, 1)
		assert.Equal(t, "subnet-id", models[0].ID.ValueString())
		assert.Equal(t, "test-subnet", models[0].Name.ValueString())
		assert.Equal(t, "10.0.0.0/25", models[0].CIDR.ValueString())

		var tagsMap map[string]string
		require.False(t, models[0].Tags.ElementsAs(ctx, &tagsMap, false).HasError())
		assert.Equal(t, map[string]string{"k1": "v1"}, tagsMap)
	})

	t.Run("empty result produces null subnets list", func(t *testing.T) {
		fc := &mockDataClient{response: []client.Subnet{}}
		d.client = fc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{Config: buildConfig("empty-vpc")}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out DataSourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		assert.True(t, out.Subnets.IsNull())
	})

	t.Run("client error becomes diagnostic error", func(t *testing.T) {
		fc := &mockDataClient{err: assert.AnError}
		d.client = fc

		resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
		d.Read(ctx, datasource.ReadRequest{Config: buildConfig("test-vpc")}, resp)
		assert.True(t, resp.Diagnostics.HasError())
	})
}
