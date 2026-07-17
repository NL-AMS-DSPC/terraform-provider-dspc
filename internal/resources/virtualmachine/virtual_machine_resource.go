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
	CreateVM(ctx context.Context, createVMRequest client.CreateVMRequest) (client.VM, error)
	DeleteVM(ctx context.Context, name string) error
	GetVM(ctx context.Context, name string) (client.VM, error)
	ListVMs(ctx context.Context) ([]client.VM, error)
}

// VMResource defines the resource implementation.
type VMResource struct {
	client VMResourceClient
}

// CPUScalerModel describes the CPU scaler configuration data model.
type CPUScalerModel struct {
	TargetUtilizationPercentage types.Int64 `tfsdk:"target_utilization_percentage"`
}

// MemoryScalerModel describes the Memory scaler configuration data model.
type MemoryScalerModel struct {
	TargetUtilizationPercentage types.Int64 `tfsdk:"target_utilization_percentage"`
}

// CronScalerModel describes the Cron scaler configuration data model.
type CronScalerModel struct {
	Timezone        types.String `tfsdk:"timezone"`
	Start           types.String `tfsdk:"start"`
	End             types.String `tfsdk:"end"`
	DesiredReplicas types.Int32  `tfsdk:"desired_replicas"`
}

// ScaleToZeroScalerModel describes the ScaleToZero scaler configuration data model.
type ScaleToZeroScalerModel struct {
	Enabled           types.Bool  `tfsdk:"enabled"`
	IdleReplicaCount  types.Int32 `tfsdk:"idle_replica_count"`
	CooldownPeriodSec types.Int32 `tfsdk:"cooldown_period_sec"`
}

// ScalersModel describes the scalers configuration data model.
type ScalersModel struct {
	CPU         *CPUScalerModel         `tfsdk:"cpu"`
	Memory      *MemoryScalerModel      `tfsdk:"memory"`
	Cron        *CronScalerModel        `tfsdk:"cron"`
	ScaleToZero *ScaleToZeroScalerModel `tfsdk:"scale_to_zero"`
}

// AutoscalingConfigModel describes the autoscaling configuration data model.
type AutoscalingConfigModel struct {
	MinReplicas types.Int64   `tfsdk:"min_replicas"`
	MaxReplicas types.Int64   `tfsdk:"max_replicas"`
	Scalers     *ScalersModel `tfsdk:"scalers"`
}

// VMResourceModel describes the resource data model.
type VMResourceModel struct {
	ID          types.String            `tfsdk:"id"`
	Name        types.String            `tfsdk:"name"`
	SkuID       types.String            `tfsdk:"sku_id"`
	Autoscaling *AutoscalingConfigModel `tfsdk:"autoscaling"`
	Replicas    types.Int32             `tfsdk:"replicas"`
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
			"replicas": schema.Int32Attribute{
				Description: "The current number of VM replicas.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"autoscaling": schema.SingleNestedBlock{
				Description: "Autoscaling configuration for the VM. When configured, the VM will automatically scale based on configured scalers.",
				Attributes: map[string]schema.Attribute{
					"min_replicas": schema.Int64Attribute{
						Description: "Minimum number of VM replicas (1-100). Defaults to 1.",
						Optional:    true,
					},
					"max_replicas": schema.Int64Attribute{
						Description: "Maximum number of VM replicas (1-100). Defaults to 1.",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"scalers": schema.SingleNestedBlock{
						Description: "Collection of scaler configurations for autoscaling.",
						Blocks: map[string]schema.Block{
							"cpu": schema.SingleNestedBlock{
								Description: "CPU-based horizontal pod autoscaling configuration.",
								Attributes: map[string]schema.Attribute{
									"target_utilization_percentage": schema.Int64Attribute{
										Description: "Target CPU utilization percentage (1-100). The VM will scale when average CPU exceeds this threshold.",
										Optional:    true,
									},
								},
							},
							"memory": schema.SingleNestedBlock{
								Description: "Memory-based horizontal pod autoscaling configuration.",
								Attributes: map[string]schema.Attribute{
									"target_utilization_percentage": schema.Int64Attribute{
										Description: "Target memory utilization percentage (1-100). The VM will scale when average memory exceeds this threshold.",
										Optional:    true,
									},
								},
							},
							"cron": schema.SingleNestedBlock{
								Description: "Cron-based scheduling configuration for scaling.",
								Attributes: map[string]schema.Attribute{
									"timezone": schema.StringAttribute{
										Description: "Timezone for cron schedule (e.g., \"Europe/Amsterdam\").",
										Optional:    true,
									},
									"start": schema.StringAttribute{
										Description: "Cron expression for scaling up (e.g., \"0 8 * * 1-5\").",
										Optional:    true,
									},
									"end": schema.StringAttribute{
										Description: "Cron expression for scaling down (e.g., \"0 18 * * 1-5\").",
										Optional:    true,
									},
									"desired_replicas": schema.Int32Attribute{
										Description: "Target replicas during active period.",
										Optional:    true,
									},
								},
							},
							"scale_to_zero": schema.SingleNestedBlock{
								Description: "Scale-to-zero configuration (KEDA-based).",
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Description: "Enable KEDA-based scale-to-zero functionality.",
										Optional:    true,
									},
									"idle_replica_count": schema.Int32Attribute{
										Description: "Number of replicas to maintain during idle (typically 0).",
										Optional:    true,
									},
									"cooldown_period_sec": schema.Int32Attribute{
										Description: "Seconds of inactivity before scaling to zero (60-3600).",
										Optional:    true,
									},
								},
							},
						},
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

		// Convert scalers if present
		if plan.Autoscaling.Scalers != nil {
			autoscaling.Scalers = &client.Scalers{}

			// CPU Scaler
			if plan.Autoscaling.Scalers.CPU != nil {
				autoscaling.Scalers.CPU = &client.CPUScaler{}
				if !plan.Autoscaling.Scalers.CPU.TargetUtilizationPercentage.IsNull() {
					autoscaling.Scalers.CPU.TargetUtilizationPercentage = plan.Autoscaling.Scalers.CPU.TargetUtilizationPercentage.ValueInt64Pointer()
				}
			}

			// Memory Scaler
			if plan.Autoscaling.Scalers.Memory != nil {
				autoscaling.Scalers.Memory = &client.MemoryScaler{}
				if !plan.Autoscaling.Scalers.Memory.TargetUtilizationPercentage.IsNull() {
					autoscaling.Scalers.Memory.TargetUtilizationPercentage = plan.Autoscaling.Scalers.Memory.TargetUtilizationPercentage.ValueInt64Pointer()
				}
			}

			// Cron Scaler
			if plan.Autoscaling.Scalers.Cron != nil {
				autoscaling.Scalers.Cron = &client.CronScaler{}
				if !plan.Autoscaling.Scalers.Cron.Timezone.IsNull() {
					autoscaling.Scalers.Cron.Timezone = plan.Autoscaling.Scalers.Cron.Timezone.ValueString()
				}
				if !plan.Autoscaling.Scalers.Cron.Start.IsNull() {
					autoscaling.Scalers.Cron.Start = plan.Autoscaling.Scalers.Cron.Start.ValueString()
				}
				if !plan.Autoscaling.Scalers.Cron.End.IsNull() {
					autoscaling.Scalers.Cron.End = plan.Autoscaling.Scalers.Cron.End.ValueString()
				}
				if !plan.Autoscaling.Scalers.Cron.DesiredReplicas.IsNull() {
					val := plan.Autoscaling.Scalers.Cron.DesiredReplicas.ValueInt32()
					autoscaling.Scalers.Cron.DesiredReplicas = &val
				}
			}

			// ScaleToZero Scaler
			if plan.Autoscaling.Scalers.ScaleToZero != nil {
				autoscaling.Scalers.ScaleToZero = &client.ScaleToZeroScaler{}
				if !plan.Autoscaling.Scalers.ScaleToZero.Enabled.IsNull() {
					autoscaling.Scalers.ScaleToZero.Enabled = plan.Autoscaling.Scalers.ScaleToZero.Enabled.ValueBoolPointer()
				}
				if !plan.Autoscaling.Scalers.ScaleToZero.IdleReplicaCount.IsNull() {
					val := plan.Autoscaling.Scalers.ScaleToZero.IdleReplicaCount.ValueInt32()
					autoscaling.Scalers.ScaleToZero.IdleReplicaCount = &val
				}
				if !plan.Autoscaling.Scalers.ScaleToZero.CooldownPeriodSec.IsNull() {
					val := plan.Autoscaling.Scalers.ScaleToZero.CooldownPeriodSec.ValueInt32()
					autoscaling.Scalers.ScaleToZero.CooldownPeriodSec = &val
				}
			}
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
		plan.Replicas = types.Int32Value(*vm.Replicas)
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
		state.Replicas = types.Int32Value(*vm.Replicas)
	}

	// Convert autoscaling config from client model to Terraform model
	if vm.Autoscaling != nil {
		state.Autoscaling = &AutoscalingConfigModel{}

		if vm.Autoscaling.MinReplicas != nil {
			state.Autoscaling.MinReplicas = types.Int64Value(*vm.Autoscaling.MinReplicas)
		} else {
			state.Autoscaling.MinReplicas = types.Int64Null()
		}

		if vm.Autoscaling.MaxReplicas != nil {
			state.Autoscaling.MaxReplicas = types.Int64Value(*vm.Autoscaling.MaxReplicas)
		} else {
			state.Autoscaling.MaxReplicas = types.Int64Null()
		}

		// Convert scalers if present
		if vm.Autoscaling.HasScalers() {
			state.Autoscaling.Scalers = &ScalersModel{}

			// CPU Scaler
			if vm.Autoscaling.Scalers.HasCPUScaler() {
				state.Autoscaling.Scalers.CPU = &CPUScalerModel{}
				if vm.Autoscaling.Scalers.CPU.TargetUtilizationPercentage != nil {
					state.Autoscaling.Scalers.CPU.TargetUtilizationPercentage = types.Int64Value(*vm.Autoscaling.Scalers.CPU.TargetUtilizationPercentage)
				} else {
					state.Autoscaling.Scalers.CPU.TargetUtilizationPercentage = types.Int64Null()
				}
			}

			// Memory Scaler
			if vm.Autoscaling.Scalers.HasMemoryScaler() {
				state.Autoscaling.Scalers.Memory = &MemoryScalerModel{}
				if vm.Autoscaling.Scalers.Memory.TargetUtilizationPercentage != nil {
					state.Autoscaling.Scalers.Memory.TargetUtilizationPercentage = types.Int64Value(*vm.Autoscaling.Scalers.Memory.TargetUtilizationPercentage)
				} else {
					state.Autoscaling.Scalers.Memory.TargetUtilizationPercentage = types.Int64Null()
				}
			}

			// Cron Scaler
			if vm.Autoscaling.Scalers.HasCronScaler() {
				state.Autoscaling.Scalers.Cron = &CronScalerModel{}
				state.Autoscaling.Scalers.Cron.Timezone = types.StringValue(vm.Autoscaling.Scalers.Cron.Timezone)
				state.Autoscaling.Scalers.Cron.Start = types.StringValue(vm.Autoscaling.Scalers.Cron.Start)
				state.Autoscaling.Scalers.Cron.End = types.StringValue(vm.Autoscaling.Scalers.Cron.End)
				if vm.Autoscaling.Scalers.Cron.DesiredReplicas != nil {
					state.Autoscaling.Scalers.Cron.DesiredReplicas = types.Int32Value(*vm.Autoscaling.Scalers.Cron.DesiredReplicas)
				} else {
					state.Autoscaling.Scalers.Cron.DesiredReplicas = types.Int32Null()
				}
			}

			// ScaleToZero Scaler
			if vm.Autoscaling.Scalers.HasScaleToZeroScaler() {
				state.Autoscaling.Scalers.ScaleToZero = &ScaleToZeroScalerModel{}
				if vm.Autoscaling.Scalers.ScaleToZero.Enabled != nil {
					state.Autoscaling.Scalers.ScaleToZero.Enabled = types.BoolValue(*vm.Autoscaling.Scalers.ScaleToZero.Enabled)
				} else {
					state.Autoscaling.Scalers.ScaleToZero.Enabled = types.BoolNull()
				}
				if vm.Autoscaling.Scalers.ScaleToZero.IdleReplicaCount != nil {
					state.Autoscaling.Scalers.ScaleToZero.IdleReplicaCount = types.Int32Value(*vm.Autoscaling.Scalers.ScaleToZero.IdleReplicaCount)
				} else {
					state.Autoscaling.Scalers.ScaleToZero.IdleReplicaCount = types.Int32Null()
				}
				if vm.Autoscaling.Scalers.ScaleToZero.CooldownPeriodSec != nil {
					state.Autoscaling.Scalers.ScaleToZero.CooldownPeriodSec = types.Int32Value(*vm.Autoscaling.Scalers.ScaleToZero.CooldownPeriodSec)
				} else {
					state.Autoscaling.Scalers.ScaleToZero.CooldownPeriodSec = types.Int32Null()
				}
			}
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
