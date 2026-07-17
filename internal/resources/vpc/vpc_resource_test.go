package vpc

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockResourceClient implements ResourceClient and records the arguments it was called with
type mockResourceClient struct {
	createID       string
	createName     string
	createTags     []client.Tag
	createSubnets  []client.CreateSubnetRequest
	createResponse client.CreateVPCResponse
	createErr      error

	getResponse client.VPC
	getErr      error
}

func (f *mockResourceClient) CreateVPC(_ context.Context, id, name string, tags []client.Tag, subnets []client.CreateSubnetRequest) (client.CreateVPCResponse, error) {
	f.createID = id
	f.createName = name
	f.createTags = tags
	f.createSubnets = subnets
	return f.createResponse, f.createErr
}

func (f *mockResourceClient) GetVPC(_ context.Context, _ string) (client.VPC, error) {
	return f.getResponse, f.getErr
}

func (f *mockResourceClient) DeleteVPC(_ context.Context, _ string) error {
	return nil
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	r := &Resource{}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	tagsValue, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"k1": "v1"})
	require.False(t, d.HasError())

	listType, ok := schemaResp.Schema.Attributes["subnets"].GetType().(types.ListType)
	if !ok {
		t.Errorf("failed to get ListType from schema for subnets")
	}
	subnetsElemType := listType.ElemType

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	diags := plan.Set(ctx, &ResourceModel{
		Name:    types.StringValue("test-vpc"),
		Tags:    tagsValue,
		Subnets: types.ListNull(subnetsElemType),
	})
	require.False(t, diags.HasError(), diags)

	fc := &mockResourceClient{
		createResponse: client.CreateVPCResponse{ID: "new-id", URN: "new-urn", Name: "test-vpc"},
		getResponse: client.VPC{
			ID:     "new-id",
			URN:    "new-urn",
			Name:   "test-vpc",
			CIDR:   "new-cidr",
			Status: "active",
			Tags:   []client.Tag{{Key: "k1", Value: "v1"}},
		},
	}
	r.client = fc

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	// assert on what the client actually received
	assert.Equal(t, "test-vpc", fc.createName)
	assert.Equal(t, []client.Tag{{Key: "k1", Value: "v1"}}, fc.createTags)
	assert.Empty(t, fc.createSubnets)

	var out ResourceModel
	require.False(t, resp.State.Get(ctx, &out).HasError())
	assert.Equal(t, "new-id", out.ID.ValueString())
	assert.Equal(t, "active", out.Status.ValueString())
}

func TestImportState(t *testing.T) {
	ctx := context.Background()
	r := &Resource{}

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	schemaType := schemaResp.Schema.Type().TerraformType(ctx)

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaType, nil)},
	}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "test-vpc"}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	var out ResourceModel
	require.False(t, resp.State.Get(ctx, &out).HasError())
	assert.Equal(t, "test-vpc", out.Name.ValueString())
}

func TestUpdate(t *testing.T) {
	vpcResource := &Resource{}

	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	vpcResource.Update(context.Background(), req, resp)

	// Update should always return an error
	if !resp.Diagnostics.HasError() {
		t.Error("Expected error from Update, got none")
	}
}
