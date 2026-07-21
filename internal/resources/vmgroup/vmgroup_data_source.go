// Package vmgroup implements the vmgroup data source.
package vmgroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/virtualmachine"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &VMGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &VMGroupDataSource{}
)

// VMGroupDataClient defines an interface for interacting with vmgroup data operations.
// ListVMs retrieves a list of virtual machines groups from the data source.
type VMGroupDataClient interface {
	ListVMGroups(ctx context.Context) ([]client.VMGroup, error)
}

// VMGroupDataSource defines the data source implementation.
type VMGroupDataSource struct {
	client VMGroupDataClient
}

// VMGroupDataSourceModel describes the data source data model.
type VMGroupDataSourceModel struct {
	VMGroups []VMGroupModel `tfsdk:"vm_groups"`
}

// VMGroupModel represents a single VMGroup in the data source
type VMGroupModel struct {
	URN               types.String            `tfsdk:"urn"`
	Name              types.String            `tfsdk:"name"`
	SKU               virtualmachine.SKUModel `tfsdk:"sku"`
	Status            types.String            `tfsdk:"status"`
	LastError         types.String            `tfsdk:"last_error"`
	Tags              types.Map               `tfsdk:"tags"`
	AttachedVolumes   []types.String          `tfsdk:"attached_volumes"`
	AutoscalingPolicy *AutoscalingPolicy      `tfsdk:"autoscaling_policy"`
}

// AutoscalingPolicy contains all available scaler configurations
type AutoscalingPolicy struct {
	MinReplicas types.Int32 `tfsdk:"min_replicas"`
	MaxReplicas types.Int32 `tfsdk:"max_replicas"`
	CronRule    *CronRule   `tfsdk:"cron"`
}

// CronRule represents cron-based scheduling configuration for scaling
type CronRule struct {
	Timezone        types.String `tfsdk:"timezone"`
	Start           types.String `tfsdk:"start"`
	End             types.String `tfsdk:"end"`
	DesiredReplicas types.Int32  `tfsdk:"desired_replicas"`
}

// NewVMGroupDataSource creates a new VMGroupDataSource.
func NewVMGroupDataSource() datasource.DataSource {
	return &VMGroupDataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *VMGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_groups"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *VMGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	cronAttrs := map[string]schema.Attribute{
		"timezone": schema.StringAttribute{
			Description: "The timezone the cron schedule is evaluated in.",
			Computed:    true,
		},
		"start": schema.StringAttribute{
			Description: "The cron expression marking the start of the scaling window.",
			Computed:    true,
		},
		"end": schema.StringAttribute{
			Description: "The cron expression marking the end of the scaling window.",
			Computed:    true,
		},
		"desired_replicas": schema.Int32Attribute{
			Description: "The desired number of replicas during the scaling window.",
			Computed:    true,
		},
	}

	autoscalingPolicyAttrs := map[string]schema.Attribute{
		"min_replicas": schema.Int32Attribute{
			Description: "The minimum number of replicas in the group.",
			Computed:    true,
		},
		"max_replicas": schema.Int32Attribute{
			Description: "The maximum number of replicas in the group.",
			Computed:    true,
		},
		"cron": schema.SingleNestedAttribute{
			Description: "Cron-based scheduling configuration for scaling.",
			Computed:    true,
			Attributes:  cronAttrs,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all virtual machine groups in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"vm_groups": schema.ListNestedAttribute{
				Description: "List of virtual machine groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"urn": schema.StringAttribute{
							Description: "The uniform resource name for the virtual machine group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the virtual machine group.",
							Computed:    true,
						},
						"sku": schema.SingleNestedAttribute{
							Description: "The SKU of the virtual machines in the group.",
							Computed:    true,
							Attributes:  virtualmachine.SKUDataSourceAttributes(),
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
							Computed:    true,
							ElementType: types.StringType,
						},
						"attached_volumes": schema.ListAttribute{
							Description: "The list of volume names attached to the virtual machine group.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"autoscaling_policy": schema.SingleNestedAttribute{
							Description: "The autoscaling policy configured for the virtual machine group.",
							Computed:    true,
							Attributes:  autoscalingPolicyAttrs,
						},
					},
				},
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *VMGroupDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.VMGroups == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected vmgroups service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.VMGroups
}

// Read reads the data from the API and stores it in the state.
func (d *VMGroupDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VMGroupDataSourceModel

	// Get all VMGroups from the API
	vms, err := d.client.ListVMGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing VMGroups",
			fmt.Sprintf("Could not list VMGroups: %s", err.Error()),
		)
		return
	}

	// Convert API VMGroups to Terraform model
	state.VMGroups = make([]VMGroupModel, len(vms))
	for i, vm := range vms {
		attachedVolumes := make([]types.String, len(vm.AttachedVolumes))
		for j, v := range vm.AttachedVolumes {
			attachedVolumes[j] = types.StringValue(v)
		}

		state.VMGroups[i] = VMGroupModel{
			URN:               types.StringValue(vm.URN),
			Name:              types.StringValue(vm.Name),
			SKU:               virtualmachine.ToTerraformSKU(vm.SKU),
			Status:            types.StringValue(vm.Status),
			LastError:         types.StringValue(vm.LastError),
			Tags:              tags.ToTerraform(ctx, vm.Tags, &resp.Diagnostics),
			AttachedVolumes:   attachedVolumes,
			AutoscalingPolicy: toTerraformAutoscaling(vm.AutoscalingPolicy),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// toTerraformAutoscaling converts a client.AutoscalingPolicy into the terraform AutoscalingPolicy model.
func toTerraformAutoscaling(policy *client.AutoscalingPolicy) *AutoscalingPolicy {
	if policy == nil {
		return nil
	}

	return &AutoscalingPolicy{
		MinReplicas: types.Int32Value(int32(policy.MinReplicas)), // nolint:gosec
		MaxReplicas: types.Int32Value(int32(policy.MaxReplicas)), // nolint:gosec
		CronRule:    toTerraformCron(policy.CronRule),
	}
}

// toTerraformCron converts a client.CronRule into the terraform CronRule model.
func toTerraformCron(rule *client.CronRule) *CronRule {
	if rule == nil {
		return nil
	}

	return &CronRule{
		Timezone:        types.StringValue(rule.Timezone),
		Start:           types.StringValue(rule.Start),
		End:             types.StringValue(rule.End),
		DesiredReplicas: types.Int32Value(int32(rule.DesiredReplicas)), // nolint:gosec
	}
}
