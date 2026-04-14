package function

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

func TestFunctionResource_updateModelFromFunction_PortHandling(t *testing.T) {
	resource := &Resource{}
	ctx := context.Background()

	t.Run("API returns zero port - should use default 8080", func(t *testing.T) {
		// Test when API returns Port = 0 (not set by service)
		function := &client.Function{
			Name:   "test-function",
			Image:  "test-image:latest",
			Port:   0, // API returns 0 (not set)
			Status: "Running",
		}

		model := &ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Null(), // Initially null (user didn't specify)
		}

		// Call updateModelFromFunction
		diags := resource.updateModelFromFunction(ctx, model, function)
		assert.False(t, diags.HasError(), "updateModelFromFunction should not return diagnostics errors")

		// Verify port is set to default 8080 when API returns 0
		assert.False(t, model.Port.IsNull(), "Port should not be null when API returns 0")
		assert.Equal(t, int64(8080), model.Port.ValueInt64(), "Port should default to 8080 when API returns 0")
	})

	t.Run("API returns specific port - should use that value", func(t *testing.T) {
		// Test when API returns a specific port
		function := &client.Function{
			Name:   "test-function",
			Image:  "test-image:latest",
			Port:   9090, // API returns specific port
			Status: "Running",
		}

		model := &ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Null(),
		}

		diags := resource.updateModelFromFunction(ctx, model, function)
		assert.False(t, diags.HasError(), "updateModelFromFunction should not return diagnostics errors")

		// Verify port is set to API value
		assert.False(t, model.Port.IsNull(), "Port should not be null after update")
		assert.False(t, model.Port.IsUnknown(), "Port should not be unknown after update")
		assert.Equal(t, int64(9090), model.Port.ValueInt64(), "Port should match API response")
	})
}

func TestFunctionResource_buildCreateFunctionRequest_PortHandling(t *testing.T) {
	resource := &Resource{}
	ctx := context.Background()

	t.Run("User specifies port - should use that value", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Value(3000), // User specified port
		}

		req, diags := resource.buildCreateFunctionRequest(ctx, plan)
		assert.False(t, diags.HasError(), "buildCreateFunctionRequest should not return diagnostics errors")

		assert.Equal(t, "test-function", req.Name)
		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(3000), req.Port, "Should use user-specified port")
	})

	t.Run("User does not specify port - should send default 8080", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Null(), // User didn't specify port
		}

		req, diags := resource.buildCreateFunctionRequest(ctx, plan)
		assert.False(t, diags.HasError(), "buildCreateFunctionRequest should not return diagnostics errors")

		assert.Equal(t, "test-function", req.Name)
		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(8080), req.Port, "Should send default 8080 when not specified to prevent null value errors")
	})
}

func TestFunctionResource_buildUpdateFunctionRequest_PortHandling(t *testing.T) {
	resource := &Resource{}
	ctx := context.Background()

	t.Run("User specifies port - should use that value", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Value(4000), // User specified port
		}

		req, diags := resource.buildUpdateFunctionRequest(ctx, plan)
		assert.False(t, diags.HasError(), "buildUpdateFunctionRequest should not return diagnostics errors")

		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(4000), req.Port, "Should use user-specified port")
	})

	t.Run("User does not specify port - should send default 8080", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Null(), // User didn't specify port
		}

		req, diags := resource.buildUpdateFunctionRequest(ctx, plan)
		assert.False(t, diags.HasError(), "buildUpdateFunctionRequest should not return diagnostics errors")

		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(8080), req.Port, "Should send default 8080 when not specified to prevent null value errors")
	})
}
