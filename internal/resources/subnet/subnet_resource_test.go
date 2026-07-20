package subnet

import (
	"context"
	"fmt"
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
	createVPCName  string
	createRequest  client.CreateSubnetRequest
	createResponse client.Subnet
	createErr      error
}

func (m *mockResourceClient) GetSubnet(_ context.Context, _, _ string) (client.Subnet, error) {
	return client.Subnet{}, nil
}

func (m *mockResourceClient) CreateSubnet(_ context.Context, vpcName string, request client.CreateSubnetRequest) (client.Subnet, error) {
	m.createVPCName = vpcName
	m.createRequest = request
	return m.createResponse, m.createErr
}

func (m *mockResourceClient) DeleteSubnet(_ context.Context, _, _ string) error {
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

	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	// populate terraform plan
	diags := plan.Set(ctx, &ResourceModel{
		Name:    types.StringValue("test-subnet"),
		VPCName: types.StringValue("test-vpc"),
		VPCID:   types.StringValue("test-vpc-id"),
		CIDR:    types.StringValue("10.0.0.0/25"),
		Type:    types.StringValue("public"),
		Tags:    tagsValue,
	})
	require.False(t, diags.HasError(), diags)

	mc := &mockResourceClient{
		createResponse: client.Subnet{ID: "new-id", URN: "new-urn", Name: "test-subnet", Status: "active"},
	}
	r.client = mc

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	// assert on what the client actually received
	assert.Equal(t, "test-vpc", mc.createVPCName)
	assert.Equal(t, "test-vpc-id", mc.createRequest.VPCID)
	assert.Equal(t, "test-subnet", mc.createRequest.Name)
	assert.Equal(t, "10.0.0.0/25", mc.createRequest.CIDR)
	assert.Equal(t, "public", mc.createRequest.Type)
	assert.Equal(t, []client.Tag{{Key: "k1", Value: "v1"}}, mc.createRequest.Tags)

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

	t.Run("valid composite ID sets vpc_name and name", func(t *testing.T) {
		resp := &resource.ImportStateResponse{
			State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaType, nil)},
		}
		r.ImportState(ctx, resource.ImportStateRequest{ID: "my-vpc:my-subnet"}, resp)
		require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

		var out ResourceModel
		require.False(t, resp.State.Get(ctx, &out).HasError())
		assert.Equal(t, "my-vpc", out.VPCName.ValueString())
		assert.Equal(t, "my-subnet", out.Name.ValueString())
	})

	t.Run("missing colon returns error", func(t *testing.T) {
		resp := &resource.ImportStateResponse{
			State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaType, nil)},
		}
		r.ImportState(ctx, resource.ImportStateRequest{ID: "my-vpc-my-subnet"}, resp)
		assert.True(t, resp.Diagnostics.HasError())
	})

	t.Run("empty ID returns error", func(t *testing.T) {
		resp := &resource.ImportStateResponse{
			State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaType, nil)},
		}
		r.ImportState(ctx, resource.ImportStateRequest{ID: ""}, resp)
		assert.True(t, resp.Diagnostics.HasError())
	})
}

func TestSplitImportID(t *testing.T) {
	tests := []struct {
		name        string
		importID    string
		expectParts int
	}{
		{
			name:        "valid composite ID",
			importID:    "my-vpc:my-subnet",
			expectParts: 2,
		},
		{
			name:        "missing colon",
			importID:    "my-vpc-my-subnet",
			expectParts: 1,
		},
		{
			name:        "empty string",
			importID:    "",
			expectParts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitImportID(tt.importID)
			assert.Len(t, parts, tt.expectParts)
		})
	}
}

func TestUpdate(t *testing.T) {
	subnetResource := &Resource{}

	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	subnetResource.Update(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected error from Update, got none")
	}
}

func TestCreateSubnetStateID(t *testing.T) {
	id := createSubnetStateID("my-vpc", "my-subnet")
	assert.Equal(t, "my-vpc:my-subnet", id)
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "not found error",
			err:      fmt.Errorf("subnet %q not found in VPC %q", "test", "vpc"),
			expected: true,
		},
		{
			name:     "API 404 error",
			err:      fmt.Errorf("API error 404: not found"),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
