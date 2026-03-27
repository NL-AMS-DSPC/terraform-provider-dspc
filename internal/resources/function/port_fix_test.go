package function

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

func TestFunctionResource_updateModelFromFunction_PortHandling(t *testing.T) {
	resource := &Resource{}

	t.Run("API returns zero port - should set null", func(t *testing.T) {
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
		resource.updateModelFromFunction(model, function)

		// Verify port is null when API returns 0 (service determines the default)
		assert.True(t, model.Port.IsNull(), "Port should be null when API returns 0")
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

		resource.updateModelFromFunction(model, function)

		// Verify port is set to API value
		assert.False(t, model.Port.IsNull(), "Port should not be null after update")
		assert.False(t, model.Port.IsUnknown(), "Port should not be unknown after update")
		assert.Equal(t, int64(9090), model.Port.ValueInt64(), "Port should match API response")
	})
}

func TestFunctionResource_buildCreateFunctionRequest_PortHandling(t *testing.T) {
	resource := &Resource{}

	t.Run("User specifies port - should use that value", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Value(3000), // User specified port
		}

		req := resource.buildCreateFunctionRequest(plan)

		assert.Equal(t, "test-function", req.Name)
		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(3000), req.Port, "Should use user-specified port")
	})

	t.Run("User does not specify port - should send 0 and let service decide", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Null(), // User didn't specify port
		}

		req := resource.buildCreateFunctionRequest(plan)

		assert.Equal(t, "test-function", req.Name)
		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(0), req.Port, "Should send 0 when not specified, letting the service determine the default")
	})
}

func TestFunctionResource_buildUpdateFunctionRequest_PortHandling(t *testing.T) {
	resource := &Resource{}

	t.Run("User specifies port - should use that value", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Value(4000), // User specified port
		}

		req := resource.buildUpdateFunctionRequest(plan)

		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(4000), req.Port, "Should use user-specified port")
	})

	t.Run("User does not specify port - should send 0 and let service decide", func(t *testing.T) {
		plan := ResourceModel{
			Name:  types.StringValue("test-function"),
			Image: types.StringValue("test-image:latest"),
			Port:  types.Int64Null(), // User didn't specify port
		}

		req := resource.buildUpdateFunctionRequest(plan)

		assert.Equal(t, "test-image:latest", req.Image)
		assert.Equal(t, int32(0), req.Port, "Should send 0 when not specified, letting the service determine the default")
	})
}
