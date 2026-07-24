package virtualmachine

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/sku"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing virtual machine resources.
// It provides methods to create, delete, retrieve, and list virtual machines.
type ResourceClient interface {
	CreateVM(ctx context.Context, createVMRequest client.CreateVMRequest) (client.VM, error)
	DeleteVM(ctx context.Context, name string) error
	GetVM(ctx context.Context, name string) (client.VM, error)
	ListVMs(ctx context.Context) ([]client.VM, error)
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel represents the VM resource
type ResourceModel struct {
	URN             types.String        `tfsdk:"urn"`
	Name            types.String        `tfsdk:"name"`
	SkuID           types.String        `tfsdk:"sku_id"`
	SKU             sku.Model           `tfsdk:"sku"`
	VPCID           types.String        `tfsdk:"vpc_id"`
	Image           types.String        `tfsdk:"image"`
	Status          types.String        `tfsdk:"status"`
	LastError       types.String        `tfsdk:"last_error"`
	Tags            types.Map           `tfsdk:"tags"`
	AttachedVolumes []types.String      `tfsdk:"attached_volumes"`
	OS              OSModel             `tfsdk:"os"`
	Customization   *CustomizationModel `tfsdk:"customization"`
	EnableLogging   types.Bool          `tfsdk:"enable_logging"`
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
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

// CustomizationResourceAttributes returns the resource schema attributes describing VM customization data.
func CustomizationResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"cloud_init": schema.SingleNestedAttribute{
			Description: "Optional cloud-init input.",
			Optional:    true,
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: map[string]schema.Attribute{
				"user_data": schema.StringAttribute{
					Description: "The cloud-init user-data content.",
					Optional:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"meta_data": schema.StringAttribute{
					Description: "The cloud-init meta-data content.",
					Optional:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
		"ignition": schema.SingleNestedAttribute{
			Description: "Optional ignition input.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"format": schema.StringAttribute{
					Description: "The format of the configuration.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"config": schema.StringAttribute{
					Description: "The configuration content.",
					Required:    true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
		},
	}
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	osAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of the OS.",
			Computed:    true,
		},
		"family": schema.StringAttribute{
			Description: "The family of the OS.",
			Computed:    true,
		},
		"distribution": schema.StringAttribute{
			Description: "The distribution of the OS.",
			Computed:    true,
		},
		"release": schema.StringAttribute{
			Description: "The release of the OS.",
			Computed:    true,
		},
		"display_name": schema.StringAttribute{
			Description: "The display name of the OS.",
			Computed:    true,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"urn": schema.StringAttribute{
				Description: "The uniform resource name for the virtual machine.",
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
			"sku": schema.SingleNestedAttribute{
				Description: "The full SKU details for the virtual machine.",
				Computed:    true,
				Attributes:  sku.ResourceAttributes(),
			},
			"vpc_id": schema.StringAttribute{
				Description: "The ID of the VPC to launch the virtual machine in.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Description: "The OS image to use for the virtual machine.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the virtual machine.",
				Computed:    true,
			},
			"last_error": schema.StringAttribute{
				Description: "The last error encountered during CRUD of the virtual machine.",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "User defined tags attached to the virtual machine.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"attached_volumes": schema.ListAttribute{
				Description: "The list of volume names attached to the virtual machine.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"os": schema.SingleNestedAttribute{
				Description: "Details about the virtual machine's OS.",
				Computed:    true,
				Attributes:  osAttrs,
			},
			"customization": schema.SingleNestedAttribute{
				Description: "Optional VM customization data.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: CustomizationResourceAttributes(),
			},
			"enable_logging": schema.BoolAttribute{
				Description: "Enable VM logging.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
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
			"Expected virtual machines service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.VirtualMachines
}

// Create creates a new virtual machine in the ASC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateVMRequest{
		Name:          plan.Name.ValueString(),
		SKUID:         plan.SkuID.ValueString(),
		VPCID:         plan.VPCID.ValueString(),
		Image:         plan.Image.ValueString(),
		Tags:          tags.ToClient(ctx, plan.Tags, &resp.Diagnostics),
		Customization: ToClientCustomization(plan.Customization),
		EnableLogging: plan.EnableLogging.ValueBool(),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.CreateVM(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VM",
			fmt.Sprintf("Could not create VM: %s", err.Error()),
		)
		return
	}

	toTerraform(ctx, &plan, vm, &resp.Diagnostics)

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

	toTerraform(ctx, &state, vm, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ToClientCustomization converts the terraform customization model into the client request payload.
func ToClientCustomization(c *CustomizationModel) *client.Customization {
	if c == nil {
		return nil
	}

	customization := &client.Customization{}

	if c.CloudInit != nil {
		customization.CloudInit = &client.CloudInitCustomization{
			UserData: c.CloudInit.UserData.ValueString(),
			MetaData: c.CloudInit.MetaData.ValueString(),
		}
	}

	if c.Ignition != nil {
		customization.Ignition = &client.IgnitionCustomization{
			Format: c.Ignition.Format.ValueString(),
			Config: c.Ignition.Config.ValueString(),
		}
	}

	return customization
}

// toTerraform transforms a client.VM into the terraform resource model.
func toTerraform(ctx context.Context, model *ResourceModel, vm client.VM, diags *diag.Diagnostics) {
	model.URN = types.StringValue(vm.URN)
	model.Name = types.StringValue(vm.Name)
	model.SkuID = types.StringValue(vm.SKU.ID)
	model.Status = types.StringValue(vm.Status)
	model.LastError = types.StringValue(vm.LastError)
	model.Tags = tags.ToTerraform(ctx, vm.Tags, diags)
	model.SKU = sku.ToTerraform(vm.SKU)
	model.OS = toOSModel(vm.OS)

	attachedVolumes := make([]types.String, len(vm.AttachedVolumes))
	for i, v := range vm.AttachedVolumes {
		attachedVolumes[i] = types.StringValue(v)
	}
	model.AttachedVolumes = attachedVolumes
}

// Update updates the virtual machine in the ASC platform.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Since the API only supports VM name and doesn't have update operations,
	// we treat any changes as requiring recreation (ForceNew)
	resp.Diagnostics.AddError(
		"Update not supported",
		"VM updates are not supported by the ASC API. Changes require VM recreation. "+
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
	err := r.client.DeleteVM(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VM",
			fmt.Sprintf("Could not delete VM: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the virtual machine in the ASC platform.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
