package vmgroup

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/sku"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/virtualmachine"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing virtual machine group resources.
// It provides methods to create, delete, retrieve, and list virtual machine groups.
type ResourceClient interface {
	CreateVMGroup(ctx context.Context, createRequest client.CreateVMGroupRequest) (client.VMGroup, error)
	DeleteVMGroup(ctx context.Context, name string) error
	GetVMGroup(ctx context.Context, name string) (client.VMGroup, error)
	ListVMGroups(ctx context.Context) ([]client.VMGroup, error)
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel represents the VM resource
type ResourceModel struct {
	URN               types.String                       `tfsdk:"urn"`
	Name              types.String                       `tfsdk:"name"`
	SkuID             types.String                       `tfsdk:"sku_id"`
	SKU               sku.Model                          `tfsdk:"sku"`
	VPCID             types.String                       `tfsdk:"vpc_id"`
	Image             types.String                       `tfsdk:"image"`
	Status            types.String                       `tfsdk:"status"`
	LastError         types.String                       `tfsdk:"last_error"`
	Tags              types.Map                          `tfsdk:"tags"`
	AttachedVolumes   []types.String                     `tfsdk:"attached_volumes"`
	AutoscalingPolicy *AutoscalingPolicy                 `tfsdk:"autoscaling_policy"`
	Customization     *virtualmachine.CustomizationModel `tfsdk:"customization"`
}

// CustomizationModel contains optional VM initialization data.
type CustomizationModel struct {
	CloudInit *CloudInitCustomization `tfsdk:"cloud_init"`
	Ignition  *IgnitionCustomization  `tfsdk:"ignition"`
}

// CloudInitCustomization contains optional cloud-init NoCloud input.
type CloudInitCustomization struct {
	UserData types.String `tfsdk:"user_data"`
	MetaData types.String `tfsdk:"meta_data"`
}

// IgnitionCustomization contains Ignition or Butane configuration input.
type IgnitionCustomization struct {
	Format types.String `tfsdk:"format"`
	Config types.String `tfsdk:"config"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_group"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	cronAttrs := map[string]schema.Attribute{
		"timezone": schema.StringAttribute{
			Description: "The timezone the cron schedule is evaluated in.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"start": schema.StringAttribute{
			Description: "The cron expression marking the start of the scaling window.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"end": schema.StringAttribute{
			Description: "The cron expression marking the end of the scaling window.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"desired_replicas": schema.Int32Attribute{
			Description: "The desired number of replicas during the scaling window.",
			Required:    true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
	}

	autoscalingPolicyAttrs := map[string]schema.Attribute{
		"min_replicas": schema.Int32Attribute{
			Description: "The minimum number of replicas in the group.",
			Required:    true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
		"max_replicas": schema.Int32Attribute{
			Description: "The maximum number of replicas in the group.",
			Required:    true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.RequiresReplace(),
			},
		},
		"cron": schema.SingleNestedAttribute{
			Description: "Cron-based scheduling configuration for scaling.",
			Optional:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: cronAttrs,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine group in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"urn": schema.StringAttribute{
				Description: "The uniform resource name for the virtual machine group.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the virtual machine group. Must be unique within the platform.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sku_id": schema.StringAttribute{
				Description: "The SKU ID defining the VMGroup size/type (e.g. \"gp-2\").",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sku": schema.SingleNestedAttribute{
				Description: "The full SKU details for the virtual machine. group",
				Computed:    true,
				Attributes:  sku.ResourceAttributes(),
			},
			"vpc_id": schema.StringAttribute{
				Description: "The ID of the VPC to launch the virtual machine group in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Description: "The OS image to use for the virtual machine group.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the virtual machine group.",
				Computed:    true,
			},
			"last_error": schema.StringAttribute{
				Description: "The last error encountered during CRUD of the virtual machine group.",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "User defined tags attached to the virtual machine group.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"attached_volumes": schema.ListAttribute{
				Description: "The list of volume names attached to the virtual machine group.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"autoscaling_policy": schema.SingleNestedAttribute{
				Description: "The autoscaling policy configured for the virtual machine group.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: autoscalingPolicyAttrs,
			},
			"customization": schema.SingleNestedAttribute{
				Description: "Optional VMGroup customization data.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: virtualmachine.CustomizationResourceAttributes(),
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the resource to use.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.AscClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AscClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.VirtualMachines == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected vmgroups service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.VMGroups
}

// Create creates a new virtual machine group in the ASC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateVMGroupRequest{
		Name:              plan.Name.ValueString(),
		SKUID:             plan.SkuID.ValueString(),
		VPCID:             plan.VPCID.ValueString(),
		Image:             plan.Image.ValueString(),
		Tags:              tags.ToClient(ctx, plan.Tags, &resp.Diagnostics),
		Customization:     virtualmachine.ToClientCustomization(plan.Customization),
		AutoscalingPolicy: toClientAutoscaling(plan.AutoscalingPolicy),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	vmg, err := r.client.CreateVMGroup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VMGroup",
			fmt.Sprintf("Could not create VMGroup: %s", err.Error()),
		)
		return
	}

	toTerraform(ctx, &plan, vmg, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Try to get the VM from the API
	vmg, err := r.client.GetVMGroup(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			// If VM not found, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting VMGroup",
			fmt.Sprintf("Could not get VMGroup: %s", err.Error()),
		)
		return
	}

	toTerraform(ctx, &state, vmg, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// toClientAutoscaling converts the terraform autoscaling policy model into the client request payload.
func toClientAutoscaling(p *AutoscalingPolicy) client.AutoscalingPolicy {
	if p == nil {
		return client.AutoscalingPolicy{}
	}

	policy := client.AutoscalingPolicy{
		MinReplicas: int(p.MinReplicas.ValueInt32()),
		MaxReplicas: int(p.MaxReplicas.ValueInt32()),
	}

	if p.CronRule != nil {
		policy.CronRule = &client.CronRule{
			Timezone:        p.CronRule.Timezone.ValueString(),
			Start:           p.CronRule.Start.ValueString(),
			End:             p.CronRule.End.ValueString(),
			DesiredReplicas: int(p.CronRule.DesiredReplicas.ValueInt32()),
		}
	}

	return policy
}

// toTerraform transforms a client.VM into the terraform resource model.
func toTerraform(ctx context.Context, model *ResourceModel, vm client.VMGroup, diags *diag.Diagnostics) {
	model.URN = types.StringValue(vm.URN)
	model.Name = types.StringValue(vm.Name)
	model.SkuID = types.StringValue(vm.SKU.ID)
	model.SKU = sku.ToTerraform(vm.SKU)
	model.Status = types.StringValue(vm.Status)
	model.LastError = types.StringValue(vm.LastError)
	model.Tags = tags.ToTerraform(ctx, vm.Tags, diags)

	attachedVolumes := make([]types.String, len(vm.AttachedVolumes))
	for i, v := range vm.AttachedVolumes {
		attachedVolumes[i] = types.StringValue(v)
	}
	model.AttachedVolumes = attachedVolumes
	model.AutoscalingPolicy = toTerraformAutoscaling(vm.AutoscalingPolicy)
}

// Update updates the virtual machine group in the ASC platform.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since the API only supports VMGroup name and doesn't have update operations,
	// we treat any changes as requiring recreation (ForceNew)
	resp.Diagnostics.AddError(
		"Update not supported",
		"VMGroup updates are not supported by the ASC API. Changes require VMGroup recreation. "+
			"Consider using lifecycle { ignore_changes = [name] } if you need to prevent replacement.",
	)
}

// Delete deletes the virtual machine in the ASC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the VM via the API
	err := r.client.DeleteVMGroup(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VMGroup",
			fmt.Sprintf("Could not delete VMGroup: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the virtual machine group in the ASC platform.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
