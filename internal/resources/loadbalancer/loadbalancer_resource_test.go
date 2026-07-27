package loadbalancer

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

// mockResourceClient implements ResourceClient and records the arguments it was called with
type mockResourceClient struct {
	createRequest  client.CreateLoadBalancerRequest
	createResponse client.CreateLoadBalancerResponse
	createErr      error
}

func (f *mockResourceClient) CreateLoadBalancer(_ context.Context, request client.CreateLoadBalancerRequest) (*client.CreateLoadBalancerResponse, error) {
	f.createRequest = request
	return &f.createResponse, f.createErr
}

func (f *mockResourceClient) DeleteLoadBalancer(_ context.Context, _ string) error {
	return nil
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	r := &Resource{}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	listenersType, ok := schemaResp.Schema.Attributes["listeners"].GetType().(types.ListType)
	if !ok {
		t.Fatal("failed to get ListType from schema for listeners")
	}
	listenersElemType := listenersType.ElemType

	backendsType, ok := schemaResp.Schema.Attributes["backends"].GetType().(types.ListType)
	if !ok {
		t.Fatal("failed to get ListType from schema for backends")
	}
	backendsElemType := backendsType.ElemType

	healthCheckType, ok := schemaResp.Schema.Attributes["health_check"].GetType().(types.ObjectType)
	if !ok {
		t.Fatal("failed to get ObjectType from schema for health_check")
	}

	listeners, d := types.ListValueFrom(ctx, listenersElemType, []listenerModel{
		{Protocol: types.StringValue("TCP"), Port: types.Int64Value(80)},
	})
	require.False(t, d.HasError(), d)

	backends, d := types.ListValueFrom(ctx, backendsElemType, []backendModel{
		{IP: types.StringValue("10.0.0.1"), Port: types.Int64Value(8080)},
	})
	require.False(t, d.HasError(), d)

	healthCheck, d := types.ObjectValueFrom(ctx, healthCheckType.AttrTypes, healthCheckModel{
		Protocol:           types.StringValue("HTTP"),
		Port:               types.Int64Value(8080),
		IntervalSeconds:    types.Int64Value(10),
		TimeoutSeconds:     types.Int64Value(5),
		HealthyThreshold:   types.Int64Value(2),
		UnhealthyThreshold: types.Int64Value(3),
	})
	require.False(t, d.HasError(), d)

	tagsValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"k1": "v1"})
	require.False(t, d.HasError())

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	diags := plan.Set(ctx, &ResourceModel{
		Name:        types.StringValue("test-lb"),
		Scheme:      types.StringValue("some-scheme"),
		SKUID:       types.StringValue("sku-1"),
		VPCID:       types.StringValue("vpc-1"),
		SubnetID:    types.StringValue("subnet-1"),
		Algorithm:   types.StringValue("round-robin"),
		Listeners:   listeners,
		Backends:    backends,
		HealthCheck: healthCheck,
		Tags:        tagsValue,
	})
	require.False(t, diags.HasError(), diags)

	mc := &mockResourceClient{
		createResponse: client.CreateLoadBalancerResponse{
			ID:   "new-id",
			Name: "test-lb",
		},
	}
	r.client = mc

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	// assert on what the client actually received
	assert.Equal(t, "test-lb", mc.createRequest.Name)
	assert.Equal(t, "some-scheme", mc.createRequest.Scheme)
	assert.Equal(t, "sku-1", mc.createRequest.SKUID)
	assert.Equal(t, "vpc-1", mc.createRequest.VPCID)
	assert.Equal(t, "subnet-1", mc.createRequest.SubnetID)
	assert.Equal(t, "round-robin", mc.createRequest.Algorithm)
	assert.Equal(t, []client.ListenerRequest{{Protocol: "TCP", Port: 80}}, mc.createRequest.Listeners)
	assert.Equal(t, []client.BackendRequest{{IP: "10.0.0.1", Port: 8080}}, mc.createRequest.Backends)
	assert.Equal(t, &client.HealthCheckRequest{
		Protocol:           "HTTP",
		Port:               8080,
		IntervalSeconds:    10,
		TimeoutSeconds:     5,
		HealthyThreshold:   2,
		UnhealthyThreshold: 3,
	}, mc.createRequest.HealthCheck)
	assert.Equal(t, []client.Tag{{Key: "k1", Value: "v1"}}, mc.createRequest.Tags)

	var out ResourceModel
	require.False(t, resp.State.Get(ctx, &out).HasError())
	assert.Equal(t, "new-id", out.ID.ValueString())
	assert.Equal(t, "test-lb", out.Name.ValueString())
}
