package virtualmachine

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

// mockVMResourceClient implements VMResourceClient and records the arguments it was called with.
type mockVMResourceClient struct {
	gotCreateVMRequest client.CreateVMRequest
	createResponse     client.VM
	createErr          error
}

func (m *mockVMResourceClient) CreateVM(_ context.Context, createVMRequest client.CreateVMRequest) (client.VM, error) {
	m.gotCreateVMRequest = createVMRequest
	return m.createResponse, m.createErr
}

func (m *mockVMResourceClient) DeleteVM(_ context.Context, _ string) error {
	return nil
}

func (m *mockVMResourceClient) GetVM(_ context.Context, _ string) (client.VM, error) {
	return client.VM{}, nil
}

func (m *mockVMResourceClient) ListVMs(_ context.Context) ([]client.VM, error) {
	return nil, nil
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	r := &VMResource{}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	tagsValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"k1": "v1"})
	require.False(t, d.HasError())

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	diags := plan.Set(ctx, &VMResourceModel{
		Name:          types.StringValue("test-vm"),
		SkuID:         types.StringValue("sku-id"),
		VPCID:         types.StringValue("test-vpc-id"),
		Image:         types.StringValue("test-image"),
		Tags:          tagsValue,
		EnableLogging: types.BoolValue(true),
	})
	require.False(t, diags.HasError(), diags)

	mc := &mockVMResourceClient{
		createResponse: client.VM{
			Name:   "test-vm",
			URN:    "test-vm-urn",
			Status: "active",
			SKU:    client.SKUResponse{ID: "sku-id", Name: "sku-name"},
		},
	}
	r.client = mc

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	// Assert on what the client actually received.
	assert.Equal(t, "test-vm", mc.gotCreateVMRequest.Name)
	assert.Equal(t, "sku-id", mc.gotCreateVMRequest.SKUID)
	assert.Equal(t, "test-vpc-id", mc.gotCreateVMRequest.VPCID)
	assert.Equal(t, "test-image", mc.gotCreateVMRequest.Image)
	assert.Equal(t, []client.Tag{{Key: "k1", Value: "v1"}}, mc.gotCreateVMRequest.Tags)
	assert.True(t, mc.gotCreateVMRequest.EnableLogging)

	var out VMResourceModel
	require.False(t, resp.State.Get(ctx, &out).HasError())
	assert.Equal(t, "test-vm-urn", out.URN.ValueString())
	assert.Equal(t, "active", out.Status.ValueString())
	assert.Equal(t, "sku-id", out.SKU.ID.ValueString())
}
