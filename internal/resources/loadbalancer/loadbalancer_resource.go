// Package loadbalancer implements the Terraform resource for managing load balancers.
package loadbalancer

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/resources/tags"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

// ResourceClient defines the interface for managing load balancer resources.
type ResourceClient interface {
	CreateLoadBalancer(ctx context.Context, request client.CreateLoadBalancerRequest) (*client.CreateLoadBalancerResponse, error)
	DeleteLoadBalancer(ctx context.Context, name string) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Scheme      types.String `tfsdk:"scheme"`
	SKUID       types.String `tfsdk:"sku_id"`
	VPCID       types.String `tfsdk:"vpc_id"`
	SubnetID    types.String `tfsdk:"subnet_id"`
	Listeners   types.List   `tfsdk:"listeners"`
	Backends    types.List   `tfsdk:"backends"`
	Algorithm   types.String `tfsdk:"algorithm"`
	HealthCheck types.Object `tfsdk:"health_check"`
	Tags        types.Map    `tfsdk:"tags"`
}

// listenerModel mirrors one listeners[] entry for ElementsAs extraction.
type listenerModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int64  `tfsdk:"port"`
}

// backendModel mirrors one backends[] entry for ElementsAs extraction.
type backendModel struct {
	IP   types.String `tfsdk:"ip"`
	Port types.Int64  `tfsdk:"port"`
}

// healthCheckModel mirrors the health_check nested attribute for ObjectAs extraction.
type healthCheckModel struct {
	Protocol           types.String `tfsdk:"protocol"`
	Port               types.Int64  `tfsdk:"port"`
	IntervalSeconds    types.Int64  `tfsdk:"interval_seconds"`
	TimeoutSeconds     types.Int64  `tfsdk:"timeout_seconds"`
	HealthyThreshold   types.Int64  `tfsdk:"healthy_threshold"`
	UnhealthyThreshold types.Int64  `tfsdk:"unhealthy_threshold"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a load balancer in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the load balancer.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the load balancer. Must be unique within the namespace.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scheme": schema.StringAttribute{
				Description: "The load balancer scheme",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sku_id": schema.StringAttribute{
				Description: "The SKU id for the load balancer.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Description: "The id of the VPC the load balancer is deployed into.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "The id of the subnet the load balancer is deployed into.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.StringAttribute{
				Description: "The load balancing algorithm",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"listeners": schema.ListNestedAttribute{
				Description: "listeners for the load balancer.",
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Description: "The listener protocol",
							Required:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The listener port.",
							Required:    true,
						},
					},
				},
			},
			"backends": schema.ListNestedAttribute{
				Description: "Backend targets that receive traffic from the load balancer.",
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Description: "The backend target IP address.",
							Required:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The backend target port.",
							Required:    true,
						},
					},
				},
			},
			"health_check": schema.SingleNestedAttribute{
				Description: "health-check configuration.",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"protocol": schema.StringAttribute{
						Description: "The health-check protocol.",
						Required:    true,
					},
					"port": schema.Int64Attribute{
						Description: "The health-check port.",
						Required:    true,
					},
					"interval_seconds": schema.Int64Attribute{
						Description: "The interval, in seconds, between health checks.",
						Required:    true,
					},
					"timeout_seconds": schema.Int64Attribute{
						Description: "The time, in seconds, to wait for a health-check response.",
						Required:    true,
					},
					"healthy_threshold": schema.Int64Attribute{
						Description: "The number of consecutive successful health checks required to mark a backend healthy.",
						Required:    true,
					},
					"unhealthy_threshold": schema.Int64Attribute{
						Description: "The number of consecutive failed health checks required to mark a backend unhealthy.",
						Required:    true,
					},
				},
			},
			"tags": schema.MapAttribute{
				Description: "User defined tags attached to the load balancer.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
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

	if dataClient.LoadBalancers == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected load balancer service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.LoadBalancers
}

// Create creates a new load balancer in the ASC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listeners := toClientListeners(ctx, plan.Listeners, &resp.Diagnostics)
	backends := toClientBackends(ctx, plan.Backends, &resp.Diagnostics)
	healthCheck := toClientHealthCheck(ctx, plan.HealthCheck, &resp.Diagnostics)
	lbTags := tags.ToClient(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	lb, err := r.client.CreateLoadBalancer(ctx, client.CreateLoadBalancerRequest{
		Name:        plan.Name.ValueString(),
		Scheme:      plan.Scheme.ValueString(),
		SKUID:       plan.SKUID.ValueString(),
		VPCID:       plan.VPCID.ValueString(),
		SubnetID:    plan.SubnetID.ValueString(),
		Listeners:   listeners,
		Backends:    backends,
		Algorithm:   plan.Algorithm.ValueString(),
		HealthCheck: healthCheck,
		Tags:        lbTags,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating load balancer",
			fmt.Sprintf("Could not create load balancer: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(lb.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read re-asserts the current state. Since the API has no GET, there is no way to detect drift or refresh other attributes here.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the load balancer in the ASC platform.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Load balancer updates are not supported by the ASC API. Changes require load balancer recreation. "+
			"Consider using lifecycle { ignore_changes = [...] } if you need to prevent replacement.",
	)
}

// Delete deletes the load balancer in the ASC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadBalancer(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting load balancer",
			fmt.Sprintf("Could not delete load balancer: %s", err.Error()),
		)
		return
	}
}

// ImportState imports the state of the load balancer from the ASC platform. But since Read cannot yet refresh state, only name is set.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// toClientListeners converts the Terraform listeners model into client listener requests.
func toClientListeners(ctx context.Context, listeners types.List, diags *diag.Diagnostics) []client.ListenerRequest {
	if listeners.IsNull() || listeners.IsUnknown() {
		return nil
	}

	var models []listenerModel
	diags.Append(listeners.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	res := make([]client.ListenerRequest, len(models))
	for i, m := range models {
		res[i] = client.ListenerRequest{
			Protocol: m.Protocol.ValueString(),
			Port:     safeUint32(m.Port.ValueInt64()),
		}
	}

	return res
}

// toClientBackends converts the Terraform backends model into client backend requests.
func toClientBackends(ctx context.Context, backends types.List, diags *diag.Diagnostics) []client.BackendRequest {
	if backends.IsNull() || backends.IsUnknown() {
		return nil
	}

	var models []backendModel
	diags.Append(backends.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	res := make([]client.BackendRequest, len(models))
	for i, m := range models {
		res[i] = client.BackendRequest{
			IP:   m.IP.ValueString(),
			Port: safeUint32(m.Port.ValueInt64()),
		}
	}

	return res
}

// toClientHealthCheck converts the Terraform health_check model into a client health-check request.
func toClientHealthCheck(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *client.HealthCheckRequest {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	var m healthCheckModel
	diags.Append(obj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &client.HealthCheckRequest{
		Protocol:           m.Protocol.ValueString(),
		Port:               safeUint32(m.Port.ValueInt64()),
		IntervalSeconds:    safeUint32(m.IntervalSeconds.ValueInt64()),
		TimeoutSeconds:     safeUint32(m.TimeoutSeconds.ValueInt64()),
		HealthyThreshold:   safeUint32(m.HealthyThreshold.ValueInt64()),
		UnhealthyThreshold: safeUint32(m.UnhealthyThreshold.ValueInt64()),
	}
}

// safeUint32 converts int64 to uint32, since the client expects unsigned ints and terraform doesn't support it.
func safeUint32(v int64) uint32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v) // #nosec
}
