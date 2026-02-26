package securityrule

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &Resource{}
	_ resource.ResourceWithConfigure = &Resource{}
)

// ResourceClient defines the interface for managing security rule resources.
type ResourceClient interface {
	GetSecurityGroup(ctx context.Context, name string) (*client.SecurityGroup, error)
	AddSecurityRules(ctx context.Context, sgName string, rules []client.AddRuleRequest) (*client.SecurityGroup, error)
	DeleteSecurityRule(ctx context.Context, sgName, direction string, index int) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	Direction         types.String `tfsdk:"direction"`
	Index             types.Int64  `tfsdk:"index"`
	Peers             types.List   `tfsdk:"peers"`
	Ports             types.List   `tfsdk:"ports"`
}

// PeerModel represents a security peer in the Terraform schema.
type PeerModel struct {
	PodSelector       types.Map    `tfsdk:"pod_selector"`
	NamespaceSelector types.Map    `tfsdk:"namespace_selector"`
	IPBlockCIDR       types.String `tfsdk:"ip_block_cidr"`
	IPBlockExcept     types.List   `tfsdk:"ip_block_except"`
}

// PortModel represents a security port in the Terraform schema.
type PortModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int64  `tfsdk:"port"`
}

// NewResource creates a new Resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Metadata updates the provided metadata with the resource type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_rule"
}

// Schema updates the resource schema with the attributes for the resource.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Security Rule within a Security Group in the DSPC platform. Each rule specifies allowed traffic by direction (ingress/egress), peers, and ports.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the Security Rule (security_group_name:direction:index).",
				Computed:    true,
			},
			"security_group_name": schema.StringAttribute{
				Description: "The name of the parent Security Group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Description: "The direction of the rule: \"ingress\" or \"egress\".",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index": schema.Int64Attribute{
				Description: "The 0-based index of the rule within its direction. Assigned by the API after creation.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"peers": schema.ListNestedBlock{
				Description: "The peers (sources for ingress, destinations for egress) for this rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pod_selector": schema.MapAttribute{
							Description: "Selects pods by label within the same namespace.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"namespace_selector": schema.MapAttribute{
							Description: "Selects namespaces by label (empty map = all namespaces).",
							Optional:    true,
							ElementType: types.StringType,
						},
						"ip_block_cidr": schema.StringAttribute{
							Description: "IP range in CIDR notation (e.g. \"10.0.0.0/24\").",
							Optional:    true,
						},
						"ip_block_except": schema.ListAttribute{
							Description: "CIDRs to exclude from the IP block range.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"ports": schema.ListNestedBlock{
				Description: "The ports allowed by this rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Description: "The network protocol: TCP, UDP, or SCTP.",
							Required:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The port number (1-65535).",
							Required:    true,
						},
					},
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

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Network
}

// Create creates a new Security Rule in the DSPC platform.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sgName := plan.SecurityGroupName.ValueString()
	direction := plan.Direction.ValueString()

	// Build the rule from plan
	rule := buildClientRule(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	addReq := []client.AddRuleRequest{
		{
			Direction: direction,
			Rule:      rule,
		},
	}

	updated, err := r.client.AddSecurityRules(ctx, sgName, addReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Security Rule",
			fmt.Sprintf("Could not add rule to Security Group %q: %s", sgName, err.Error()),
		)
		return
	}

	// The new rule is the last one in the appropriate direction
	var rules []client.SecurityRule
	if direction == "ingress" {
		rules = updated.IngressRules
	} else {
		rules = updated.EgressRules
	}

	if len(rules) == 0 {
		resp.Diagnostics.AddError(
			"Error creating Security Rule",
			"API returned no rules after creation",
		)
		return
	}

	newIndex := len(rules) - 1
	plan.Index = types.Int64Value(int64(newIndex))
	plan.ID = types.StringValue(createRuleStateID(sgName, direction, newIndex))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the data from the API and stores it in the state.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sgName := state.SecurityGroupName.ValueString()
	direction := state.Direction.ValueString()
	index := int(state.Index.ValueInt64())

	sg, err := r.client.GetSecurityGroup(ctx, sgName)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error getting Security Group",
			fmt.Sprintf("Could not get Security Group %q: %s", sgName, err.Error()),
		)
		return
	}

	var rules []client.SecurityRule
	if direction == "ingress" {
		rules = sg.IngressRules
	} else {
		rules = sg.EgressRules
	}

	if index >= len(rules) {
		// Rule no longer exists
		resp.State.RemoveResource(ctx)
		return
	}

	rule := rules[index]

	// Update state from read rule
	state.ID = types.StringValue(createRuleStateID(sgName, direction, index))
	state.Index = types.Int64Value(int64(index))

	// Update peers
	state.Peers = flattenPeers(ctx, rule.Peers)
	// Update ports
	state.Ports = flattenPorts(ctx, rule.Ports)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the Security Rule in the DSPC platform.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Security Rule updates are not supported by the DSPC API. Changes require rule recreation. "+
			"Delete the existing rule and create a new one with the desired configuration.",
	)
}

// Delete deletes the Security Rule in the DSPC platform.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecurityRule(
		ctx,
		state.SecurityGroupName.ValueString(),
		state.Direction.ValueString(),
		int(state.Index.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Security Rule",
			fmt.Sprintf("Could not delete Security Rule: %s", err.Error()),
		)
		return
	}
}

// buildClientRule converts the Terraform plan into a client.SecurityRule.
func buildClientRule(ctx context.Context, plan *ResourceModel, diags *diag.Diagnostics) client.SecurityRule {
	var rule client.SecurityRule

	// Extract peers
	if !plan.Peers.IsNull() && !plan.Peers.IsUnknown() {
		var peers []PeerModel
		d := plan.Peers.ElementsAs(ctx, &peers, false)
		diags.Append(d...)

		for _, p := range peers {
			peer := client.SecurityPeer{}

			if !p.PodSelector.IsNull() && !p.PodSelector.IsUnknown() {
				selectors := make(map[string]string)
				d := p.PodSelector.ElementsAs(ctx, &selectors, false)
				diags.Append(d...)
				peer.PodSelector = selectors
			}

			if !p.NamespaceSelector.IsNull() && !p.NamespaceSelector.IsUnknown() {
				selectors := make(map[string]string)
				d := p.NamespaceSelector.ElementsAs(ctx, &selectors, false)
				diags.Append(d...)
				peer.NamespaceSelector = selectors
			}

			if !p.IPBlockCIDR.IsNull() && !p.IPBlockCIDR.IsUnknown() {
				ipBlock := &client.IPBlock{
					CIDR: p.IPBlockCIDR.ValueString(),
				}
				if !p.IPBlockExcept.IsNull() && !p.IPBlockExcept.IsUnknown() {
					var excepts []string
					d := p.IPBlockExcept.ElementsAs(ctx, &excepts, false)
					diags.Append(d...)
					ipBlock.Except = excepts
				}
				peer.IPBlock = ipBlock
			}

			rule.Peers = append(rule.Peers, peer)
		}
	}

	// Extract ports
	if !plan.Ports.IsNull() && !plan.Ports.IsUnknown() {
		var ports []PortModel
		d := plan.Ports.ElementsAs(ctx, &ports, false)
		diags.Append(d...)

		for _, p := range ports {
			rule.Ports = append(rule.Ports, client.SecurityPort{
				Protocol: p.Protocol.ValueString(),
				Port:     int(p.Port.ValueInt64()),
			})
		}
	}

	return rule
}

// flattenPeers converts client.SecurityPeer slice to a Terraform list.
func flattenPeers(ctx context.Context, peers []client.SecurityPeer) types.List {
	if len(peers) == 0 {
		return types.ListNull(peerObjectType())
	}

	var peerVals []attr.Value
	for _, p := range peers {
		peerAttrs := map[string]attr.Value{}

		if len(p.PodSelector) > 0 {
			elems := make(map[string]attr.Value)
			for k, v := range p.PodSelector {
				elems[k] = types.StringValue(v)
			}
			peerAttrs["pod_selector"], _ = types.MapValue(types.StringType, elems)
		} else {
			peerAttrs["pod_selector"] = types.MapNull(types.StringType)
		}

		if len(p.NamespaceSelector) > 0 {
			elems := make(map[string]attr.Value)
			for k, v := range p.NamespaceSelector {
				elems[k] = types.StringValue(v)
			}
			peerAttrs["namespace_selector"], _ = types.MapValue(types.StringType, elems)
		} else {
			peerAttrs["namespace_selector"] = types.MapNull(types.StringType)
		}

		if p.IPBlock != nil {
			peerAttrs["ip_block_cidr"] = types.StringValue(p.IPBlock.CIDR)
			if len(p.IPBlock.Except) > 0 {
				excepts := make([]attr.Value, len(p.IPBlock.Except))
				for i, e := range p.IPBlock.Except {
					excepts[i] = types.StringValue(e)
				}
				peerAttrs["ip_block_except"], _ = types.ListValue(types.StringType, excepts)
			} else {
				peerAttrs["ip_block_except"] = types.ListNull(types.StringType)
			}
		} else {
			peerAttrs["ip_block_cidr"] = types.StringNull()
			peerAttrs["ip_block_except"] = types.ListNull(types.StringType)
		}

		obj, _ := types.ObjectValue(peerAttrTypes(), peerAttrs)
		peerVals = append(peerVals, obj)
	}

	list, _ := types.ListValue(peerObjectType(), peerVals)
	return list
}

// flattenPorts converts client.SecurityPort slice to a Terraform list.
func flattenPorts(ctx context.Context, ports []client.SecurityPort) types.List {
	if len(ports) == 0 {
		return types.ListNull(portObjectType())
	}

	var portVals []attr.Value
	for _, p := range ports {
		portAttrs := map[string]attr.Value{
			"protocol": types.StringValue(p.Protocol),
			"port":     types.Int64Value(int64(p.Port)),
		}
		obj, _ := types.ObjectValue(portAttrTypes(), portAttrs)
		portVals = append(portVals, obj)
	}

	list, _ := types.ListValue(portObjectType(), portVals)
	return list
}

func peerAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"pod_selector":       types.MapType{ElemType: types.StringType},
		"namespace_selector": types.MapType{ElemType: types.StringType},
		"ip_block_cidr":      types.StringType,
		"ip_block_except":    types.ListType{ElemType: types.StringType},
	}
}

func peerObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: peerAttrTypes()}
}

func portAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"protocol": types.StringType,
		"port":     types.Int64Type,
	}
}

func portObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: portAttrTypes()}
}

// createRuleStateID creates a unique identifier for the security rule resource.
func createRuleStateID(sgName, direction string, index int) string {
	return fmt.Sprintf("%s:%s:%d", sgName, direction, index)
}

// parseRuleStateID splits an import ID string.
func parseRuleStateID(id string) (sgName, direction string, index int, err error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("import ID must be in the format 'security-group-name:direction:index', got: %s", id)
		return
	}
	sgName = parts[0]
	direction = parts[1]
	index, err = strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("invalid index in import ID %q: %w", id, err)
	}
	return
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not found") ||
		len(err.Error()) > 14 && err.Error()[:14] == "API error 404:")
}
