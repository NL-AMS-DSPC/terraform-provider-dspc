package vmgroup

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVMGroupResourceClient implements ResourceClient and records the arguments it was called with.
type mockVMGroupResourceClient struct {
	gotCreateVMGroupRequest client.CreateVMGroupRequest
	createResponse          client.VMGroup
	createErr               error
}

func (m *mockVMGroupResourceClient) CreateVMGroup(_ context.Context, createRequest client.CreateVMGroupRequest) (client.VMGroup, error) {
	m.gotCreateVMGroupRequest = createRequest
	return m.createResponse, m.createErr
}

func (m *mockVMGroupResourceClient) DeleteVMGroup(_ context.Context, _ string) error {
	return nil
}

func (m *mockVMGroupResourceClient) GetVMGroup(_ context.Context, _ string) (client.VMGroup, error) {
	return client.VMGroup{}, nil
}

func (m *mockVMGroupResourceClient) ListVMGroups(_ context.Context) ([]client.VMGroup, error) {
	return nil, nil
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	r := &Resource{}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	tagsValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"k1": "v1"})
	require.False(t, d.HasError())

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	diags := plan.Set(ctx, &ResourceModel{
		Name:  types.StringValue("test-vmgroup"),
		SkuID: types.StringValue("sku-id"),
		VPCID: types.StringValue("test-vpc-id"),
		Image: types.StringValue("test-image"),
		Tags:  tagsValue,
		AutoscalingPolicy: &AutoscalingPolicy{
			MinReplicas: types.Int32Value(1),
			MaxReplicas: types.Int32Value(3),
			CronRule: &CronRule{
				Timezone:        types.StringValue("UTC"),
				Start:           types.StringValue("0 8 * * *"),
				End:             types.StringValue("0 18 * * *"),
				DesiredReplicas: types.Int32Value(2),
			},
		},
	})
	require.False(t, diags.HasError(), diags)

	mc := &mockVMGroupResourceClient{
		createResponse: client.VMGroup{
			Name:   "test-vmgroup",
			URN:    "test-vmgroup-urn",
			Status: "active",
			SKU:    client.SKUResponse{ID: "sku-id", Name: "sku-name"},
		},
	}
	r.client = mc

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	// Assert on what the client actually received.
	assert.Equal(t, "test-vmgroup", mc.gotCreateVMGroupRequest.Name)
	assert.Equal(t, "sku-id", mc.gotCreateVMGroupRequest.SKUID)
	assert.Equal(t, "test-vpc-id", mc.gotCreateVMGroupRequest.VPCID)
	assert.Equal(t, "test-image", mc.gotCreateVMGroupRequest.Image)
	assert.Equal(t, []client.Tag{{Key: "k1", Value: "v1"}}, mc.gotCreateVMGroupRequest.Tags)
	assert.Equal(t, client.AutoscalingPolicy{
		MinReplicas: 1,
		MaxReplicas: 3,
		CronRule: &client.CronRule{
			Timezone:        "UTC",
			Start:           "0 8 * * *",
			End:             "0 18 * * *",
			DesiredReplicas: 2,
		},
	}, mc.gotCreateVMGroupRequest.AutoscalingPolicy)

	var out ResourceModel
	require.False(t, resp.State.Get(ctx, &out).HasError())
	assert.Equal(t, "test-vmgroup-urn", out.URN.ValueString())
	assert.Equal(t, "active", out.Status.ValueString())
	assert.Equal(t, "sku-id", out.SKU.ID.ValueString())
}
