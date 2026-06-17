// Package cluster implements the Terraform resource for managing OpenShift clusters
// provisioned by the DSPC cluster-service.
package cluster

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient is the subset of the DSPC cluster-service client used by this resource.
type ResourceClient interface {
	CreateCluster(ctx context.Context, req client.ClusterCreateRequest) (*client.Cluster, error)
	GetCluster(ctx context.Context, name string) (*client.Cluster, error)
	PatchCluster(ctx context.Context, name string, tags []client.ClusterTag) (*client.Cluster, error)
	DeleteCluster(ctx context.Context, name string) error
}

// Resource is the Terraform resource implementation for a DSPC cluster.
type Resource struct {
	client ResourceClient
}

// ResourceModel is the Terraform resource model for a DSPC cluster.
type ResourceModel struct {
	ID            types.String `tfsdk:"id"`
	URN           types.String `tfsdk:"urn"`
	Name          types.String `tfsdk:"name"`
	Domain        types.String `tfsdk:"domain"`
	Version       types.String `tfsdk:"version"`
	Image         types.String `tfsdk:"image"`
	Status        types.String `tfsdk:"status"`
	StatusMessage types.String `tfsdk:"status_message"`
	CreatedAt     types.String `tfsdk:"created_at"`
	Tags          types.Map    `tfsdk:"tags"`
	ControlPlane  types.Object `tfsdk:"control_plane"`
	Workers       types.Object `tfsdk:"workers"`
	VPC           types.Object `tfsdk:"vpc"`
	PullSecret    types.String `tfsdk:"pull_secret"`
	SSHKey        types.String `tfsdk:"ssh_key"`
}

type nodePoolModel struct {
	Replicas types.Int64  `tfsdk:"replicas"`
	SKUID    types.String `tfsdk:"sku_id"`
}

type subnetModel struct {
	Name types.String `tfsdk:"name"`
}

type subnetsModel struct {
	Pods     subnetModel `tfsdk:"pods"`
	Services subnetModel `tfsdk:"services"`
}

type vpcModel struct {
	Name    types.String `tfsdk:"name"`
	Subnets subnetsModel `tfsdk:"subnets"`
}

// NewResource constructs a new cluster Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata sets the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// Schema describes the cluster resource's attributes.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	subnetSchema := schema.SingleNestedAttribute{
		Required: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the subnet to link.",
				Required:    true,
			},
		},
	}

	nodePoolSchema := schema.SingleNestedAttribute{
		Required: true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"replicas": schema.Int64Attribute{
				Description: "Number of nodes in the pool. Must be at least 1.",
				Required:    true,
			},
			"sku_id": schema.StringAttribute{
				Description: "SKU id sizing the nodes in this pool.",
				Required:    true,
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages an OpenShift cluster provisioned by the DSPC cluster-service. " +
			"Provisioning is asynchronous: cluster-service returns immediately after persisting the cluster " +
			"and continues to install the cluster in the background.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Cluster identifier (currently the cluster name).",
				Computed:    true,
			},
			"urn": schema.StringAttribute{
				Description: "URN of the cluster assigned by cluster-service.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Cluster name. Must be unique within the tenant.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Cluster base domain (e.g. \"example.com\").",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "OpenShift version to install (e.g. \"4.16.5\").",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Description: "VM image used to boot the cluster nodes.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Cluster lifecycle status (e.g. \"provisioning\", \"installing\", \"ready\", \"error\").",
				Computed:    true,
			},
			"status_message": schema.StringAttribute{
				Description: "Human-readable detail about the current status.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Cluster creation timestamp.",
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Customer-managed key/value tags.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"control_plane": nodePoolSchema,
			"workers":       nodePoolSchema,
			"vpc": schema.SingleNestedAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Name of the VPC to attach the cluster to.",
						Required:    true,
					},
					"subnets": schema.SingleNestedAttribute{
						Required: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
						Attributes: map[string]schema.Attribute{
							"pods":     subnetSchema,
							"services": subnetSchema,
						},
					},
				},
			},
			"pull_secret": schema.StringAttribute{
				Description: "Pull secret used to render the install-config. Write-only: never stored in state (Terraform >= 1.11).",
				Required:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
			"ssh_key": schema.StringAttribute{
				Description: "SSH public key authorized on the cluster nodes. Write-only: never stored in state (Terraform >= 1.11).",
				Required:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
		},
	}
}

// Configure wires the cluster-service client into the resource.
func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dspcClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dspcClient.Clusters == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected cluster service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dspcClient.Clusters
}

// Create provisions a new cluster.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// WriteOnly attributes are null in plan/state — read them from config.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlPlane, ok := nodePoolFromObject(ctx, plan.ControlPlane, &resp.Diagnostics)
	if !ok {
		return
	}
	workers, ok := nodePoolFromObject(ctx, plan.Workers, &resp.Diagnostics)
	if !ok {
		return
	}
	vpc, ok := vpcFromObject(ctx, plan.VPC, &resp.Diagnostics)
	if !ok {
		return
	}

	createReq := client.ClusterCreateRequest{
		Name:    plan.Name.ValueString(),
		Domain:  plan.Domain.ValueString(),
		Version: plan.Version.ValueString(),
		Image:   plan.Image.ValueString(),
		ControlPlane: client.ClusterNodePoolRequest{
			Replicas: safeInt32(controlPlane.Replicas.ValueInt64()),
			SKUID:    controlPlane.SKUID.ValueString(),
		},
		Workers: client.ClusterNodePoolRequest{
			Replicas: safeInt32(workers.Replicas.ValueInt64()),
			SKUID:    workers.SKUID.ValueString(),
		},
		VPC: client.ClusterVPCRequest{
			Name: vpc.Name.ValueString(),
			Subnets: client.ClusterSubnetsRequest{
				Pods:     client.ClusterSubnetRequest{Name: vpc.Subnets.Pods.Name.ValueString()},
				Services: client.ClusterSubnetRequest{Name: vpc.Subnets.Services.Name.ValueString()},
			},
		},
		PullSecret: config.PullSecret.ValueString(),
		SSHKey:     config.SSHKey.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tagsMap map[string]string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range tagsMap {
			createReq.Tags = append(createReq.Tags, client.ClusterTag{Key: k, Value: v})
		}
	}

	cluster, err := r.client.CreateCluster(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			fmt.Sprintf("Could not create cluster: %s", err.Error()),
		)
		return
	}

	mapStateFromCluster(ctx, &plan, cluster, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the cluster state from cluster-service.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetCluster(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading cluster",
			fmt.Sprintf("Could not read cluster %q: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	mapStateFromCluster(ctx, &state, cluster, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update applies in-place changes. Only tag changes are supported today;
// other fields are RequiresReplace and never reach Update.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := make([]client.ClusterTag, 0)
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tagsMap map[string]string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagsMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range tagsMap {
			tags = append(tags, client.ClusterTag{Key: k, Value: v})
		}
	}

	cluster, err := r.client.PatchCluster(ctx, state.Name.ValueString(), tags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cluster",
			fmt.Sprintf("Could not patch cluster %q: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	mapStateFromCluster(ctx, &plan, cluster, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete tears down the cluster.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCluster(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting cluster",
			fmt.Sprintf("Could not delete cluster %q: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState supports `terraform import dspc_cluster.foo <cluster-name>`.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func nodePoolFromObject(ctx context.Context, obj types.Object, diags *diag.Diagnostics) (nodePoolModel, bool) {
	var np nodePoolModel
	diags.Append(obj.As(ctx, &np, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nodePoolModel{}, false
	}
	return np, true
}

func vpcFromObject(ctx context.Context, obj types.Object, diags *diag.Diagnostics) (vpcModel, bool) {
	var v vpcModel
	diags.Append(obj.As(ctx, &v, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return vpcModel{}, false
	}
	return v, true
}

var (
	nodePoolAttrTypes = map[string]attr.Type{
		"replicas": types.Int64Type,
		"sku_id":   types.StringType,
	}
	subnetAttrTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	subnetsAttrTypes = map[string]attr.Type{
		"pods":     types.ObjectType{AttrTypes: subnetAttrTypes},
		"services": types.ObjectType{AttrTypes: subnetAttrTypes},
	}
	vpcAttrTypes = map[string]attr.Type{
		"name":    types.StringType,
		"subnets": types.ObjectType{AttrTypes: subnetsAttrTypes},
	}
)

// mapStateFromCluster maps a Cluster API response into the Terraform ResourceModel.
func mapStateFromCluster(ctx context.Context, model *ResourceModel, c *client.Cluster, diags *diag.Diagnostics) {
	model.ID = types.StringValue(c.Name)
	model.URN = types.StringValue(c.URN)
	model.Name = types.StringValue(c.Name)
	model.Domain = types.StringValue(c.Domain)
	model.Version = types.StringValue(c.Version)
	model.Image = types.StringValue(c.Image)
	model.Status = types.StringValue(c.Status)
	if c.StatusMessage != "" {
		model.StatusMessage = types.StringValue(c.StatusMessage)
	} else {
		model.StatusMessage = types.StringNull()
	}
	model.CreatedAt = types.StringValue(c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

	cpObj, d := nodePoolToObject(ctx, c.ControlPlane)
	diags.Append(d...)
	model.ControlPlane = cpObj

	wObj, d := nodePoolToObject(ctx, c.Workers)
	diags.Append(d...)
	model.Workers = wObj

	vpcObj, d := vpcToObject(ctx, c.VPC)
	diags.Append(d...)
	model.VPC = vpcObj

	if len(c.Tags) > 0 {
		tagsMap := make(map[string]string, len(c.Tags))
		for _, t := range c.Tags {
			tagsMap[t.Key] = t.Value
		}
		tagsValue, d := types.MapValueFrom(ctx, types.StringType, tagsMap)
		diags.Append(d...)
		model.Tags = tagsValue
	} else {
		model.Tags = types.MapNull(types.StringType)
	}
}

func nodePoolToObject(ctx context.Context, np client.ClusterNodePool) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, nodePoolAttrTypes, nodePoolModel{
		Replicas: types.Int64Value(int64(np.Replicas)),
		SKUID:    types.StringValue(np.SKUID),
	})
}

func vpcToObject(ctx context.Context, v client.ClusterVPC) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, vpcAttrTypes, vpcModel{
		Name: types.StringValue(v.Name),
		Subnets: subnetsModel{
			Pods:     subnetModel{Name: types.StringValue(v.Subnets.Pods.Name)},
			Services: subnetModel{Name: types.StringValue(v.Subnets.Services.Name)},
		},
	})
}

func safeInt32(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(v) // #nosec G115 -- overflow is guarded by the check above
}
