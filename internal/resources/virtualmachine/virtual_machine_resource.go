package virtualmachine

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &VMResource{}
	_ resource.ResourceWithConfigure   = &VMResource{}
	_ resource.ResourceWithImportState = &VMResource{}
)

// VMResourceClient defines the interface for managing virtual machine resources.
// It provides methods to create, delete, retrieve, and list virtual machines.
type VMResourceClient interface {
	CreateVM(ctx context.Context, name, skuID string, autoscaling *client.AutoscalingConfig) (*client.VM, error)
	DeleteVM(ctx context.Context, name string) error
	GetVM(ctx context.Context, name string) (*client.VM, error)
	ListVMs(ctx context.Context) ([]*client.VM, error)
}

// VMResource defines the resource implementation.
type VMResource struct {
	client VMResourceClient
}

// AutoscalingConfigModel describes the autoscaling configuration data model.
type AutoscalingConfigModel struct {
	MinReplicas                       types.Int64 `tfsdk:"min_replicas"`
	MaxReplicas                       types.Int64 `tfsdk:"max_replicas"`
	TargetCPUUtilizationPercentage    types.Int64 `tfsdk:"target_cpu_utilization_percentage"`
	TargetMemoryUtilizationPercentage types.Int64 `tfsdk:"target_memory_utilization_percentage"`
	EnableScaleToZero                 types.Bool  `tfsdk:"enable_scale_to_zero"`
	ScaleToZeroAfter                  types.Int64 `tfsdk:"scale_to_zero_after"`
}

// VMResourceModel describes the resource data model.
type VMResourceModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	SkuID       types.String            `tfsdk:"sku_id"`
	Autoscaling *AutoscalingConfigModel `tfsdk:"autoscaling"`
	Replicas    types.Int64             `tfsdk:"replicas"`
	Status      types.String            `tfsdk:"status"`
}

// NewVMResource creates a new VMResource.
func NewVMResource() resource.Resource {
	return &VMResource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *VMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *VMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine in the DSPC platform with optional autoscaling configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the virtual machine.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the virtual machine. Must be unique within the platform.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sku_id": schema.StringAttribute{
				Description: "The SKU ID defining the VM size/type (e.g. \"gp-2\").",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the virtual machine (e.g., \"pending\", \"ready\").",
				Computed:    true,
			},
			"replicas": schema.Int64Attribute{
				Description: "The current number of VM replicas.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"autoscaling": schema.SingleNestedBlock{
				Description: "Autoscaling configuration for the VM. When configured, the VM will automatically scale based on CPU/memory usage.",
				Attributes: map[string]schema.Attribute{
					"min_replicas": schema.Int64Attribute{
						Description: "Minimum number of VM replicas (1-100). Defaults to 1.",
						Optional:    true,
					},
					"max_replicas": schema.Int64Attribute{
						Description: "Maximum number of VM replicas (1-100). Defaults to 1.",
						Optional:    true,
					},
					"target_cpu_utilization_percentage": schema.Int64Attribute{
						Description: "Target CPU utilization percentage (1-100). The VM will scale when average CPU exceeds this threshold.",
						Optional:    true,
					},
					"target_memory_utilization_percentage": schema.Int64Attribute{
						Description: "Target memory utilization percentage (1-100). The VM will scale when average memory exceeds this threshold.",
						Optional:    true,
					},
					"enable_scale_to_zero": schema.BoolAttribute{
						Description: "Enable KEDA-based scale-to-zero functionality. When true, the VM can scale down to 0 replicas during idle periods.",
						Optional:    true,
					},
					"scale_to_zero_after": schema.Int64Attribute{
						Description: "Seconds of inactivity before scaling to zero (60-3600). Only applies when enable_scale_to_zero is true.",
						Optional:    true,
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *VMResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.VirtualMachines == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected virtual machines service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.VirtualMachines
}

// Create creates a new virtual machine in the DSPC platform.
func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VMResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Convert autoscaling config from Terraform model to client model
	var autoscaling *client.AutoscalingConfig
	if plan.Autoscaling != nil {
		autoscaling = &client.AutoscalingConfig{}

		if !plan.Autoscaling.MinReplicas.IsNull() {
			autoscaling.MinReplicas = plan.Autoscaling.MinReplicas.ValueInt64Pointer()
		}
		if !plan.Autoscaling.MaxReplicas.IsNull() {
			autoscaling.MaxReplicas = plan.Autoscaling.MaxReplicas.ValueInt64Pointer()
		}
		if !plan.Autoscaling.TargetCPUUtilizationPercentage.IsNull() {
			autoscaling.TargetCPUUtilizationPercentage = plan.Autoscaling.TargetCPUUtilizationPercentage.ValueInt64Pointer()
		}
		if !plan.Autoscaling.TargetMemoryUtilizationPercentage.IsNull() {
			autoscaling.TargetMemoryUtilizationPercentage = plan.Autoscaling.TargetMemoryUtilizationPercentage.ValueInt64Pointer()
		}
		if !plan.Autoscaling.EnableScaleToZero.IsNull() {
			autoscaling.EnableScaleToZero = plan.Autoscaling.EnableScaleToZero.ValueBoolPointer()
		}
		if !plan.Autoscaling.ScaleToZeroAfter.IsNull() {
			autoscaling.ScaleToZeroAfter = plan.Autoscaling.ScaleToZeroAfter.ValueInt64Pointer()
		}
	}

	// Create the VM via the API
	vm, err := r.client.CreateVM(ctx, plan.Name.ValueString(), plan.SkuID.ValueString(), autoscaling)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VM",
			fmt.Sprintf("Could not create VM: %s", err.Error()),
		)
		return
	}

	// Set the computed values
	plan.ID = types.StringValue(vm.Name)
	plan.SkuID = types.StringValue(vm.SKU.ID)
	plan.Status = types.StringValue(vm.Status)

	if vm.Replicas != nil {
		plan.Replicas = types.Int64Value(int64(*vm.Replicas))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the VM from the API
	vm, err := r.client.GetVM(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			// If VM not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting VM",
			fmt.Sprintf("Could not get VM: %s", err.Error()),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(vm.Name)
	state.Name = types.StringValue(vm.Name)
	state.SkuID = types.StringValue(vm.SKU.ID)
	state.Status = types.StringValue(vm.Status)

	if vm.Replicas != nil {
		state.Replicas = types.Int64Value(int64(*vm.Replicas))
	}

	// Convert autoscaling config from client model to Terraform model
	if vm.Autoscaling != nil {
		state.Autoscaling = &AutoscalingConfigModel{}

		if vm.Autoscaling.MinReplicas != nil {
			state.Autoscaling.MinReplicas = types.Int64Value(int64(*vm.Autoscaling.MinReplicas))
		} else {
			state.Autoscaling.MinReplicas = types.Int64Null()
		}

		if vm.Autoscaling.MaxReplicas != nil {
			state.Autoscaling.MaxReplicas = types.Int64Value(int64(*vm.Autoscaling.MaxReplicas))
		} else {
			state.Autoscaling.MaxReplicas = types.Int64Null()
		}

		if vm.Autoscaling.TargetCPUUtilizationPercentage != nil {
			state.Autoscaling.TargetCPUUtilizationPercentage = types.Int64Value(int64(*vm.Autoscaling.TargetCPUUtilizationPercentage))
		} else {
			state.Autoscaling.TargetCPUUtilizationPercentage = types.Int64Null()
		}

		if vm.Autoscaling.TargetMemoryUtilizationPercentage != nil {
			state.Autoscaling.TargetMemoryUtilizationPercentage = types.Int64Value(int64(*vm.Autoscaling.TargetMemoryUtilizationPercentage))
		} else {
			state.Autoscaling.TargetMemoryUtilizationPercentage = types.Int64Null()
		}

		if vm.Autoscaling.EnableScaleToZero != nil {
			state.Autoscaling.EnableScaleToZero = types.BoolValue(*vm.Autoscaling.EnableScaleToZero)
		} else {
			state.Autoscaling.EnableScaleToZero = types.BoolNull()
		}

		if vm.Autoscaling.ScaleToZeroAfter != nil {
			state.Autoscaling.ScaleToZeroAfter = types.Int64Value(int64(*vm.Autoscaling.ScaleToZeroAfter))
		} else {
			state.Autoscaling.ScaleToZeroAfter = types.Int64Null()
		}
	} else {
		state.Autoscaling = nil
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the virtual machine in the DSPC platform.
func (r *VMResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since the API only supports VM name and doesn't have update operations,
	// we treat any changes as requiring recreation (ForceNew)
	resp.Diagnostics.AddError(
		"Update not supported",
		"VM updates are not supported by the DSPC API. Changes require VM recreation. "+
			"Consider using lifecycle { ignore_changes = [name] } if you need to prevent replacement.",
	)
}

// Delete deletes the virtual machine in the DSPC platform.
func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the VM via the API
	err := r.client.DeleteVM(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VM",
			fmt.Sprintf("Could not delete VM: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the virtual machine in the DSPC platform.
func (r *VMResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
