// Package securityrule provides Terraform resources and data sources for managing security rules.
package securityrule

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &Resource{}
	_ resource.ResourceWithConfigure = &Resource{}
)

// ResourceClient defines the interface for managing security rule resources.
type ResourceClient interface {
	AddSecurityRules(ctx context.Context, sgName string, req client.AddRulesRequest) (*client.SecurityGroup, error)
	GetSecurityGroup(ctx context.Context, name string) (*client.SecurityGroup, error)
	DeleteSecurityRule(ctx context.Context, sgName, direction string, index int) error
}

// Resource defines the resource implementation.
type Resource struct {
	client ResourceClient
}

// PortModel represents a port in the resource model.
type PortModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int64  `tfsdk:"port"`
}

// PeerIPBlockModel represents an IP block in a peer.
type PeerIPBlockModel struct {
	CIDR   types.String `tfsdk:"cidr"`
	Except types.List   `tfsdk:"except"`
}

// PeerModel represents a peer in the resource model.
type PeerModel struct {
	PodSelector       types.Map         `tfsdk:"pod_selector"`
	NamespaceSelector types.Map         `tfsdk:"namespace_selector"`
	IPBlock           *PeerIPBlockModel `tfsdk:"ip_block"`
}

// ResourceModel describes the resource data model.
type ResourceModel struct {
	ID                types.String `tfsdk:"id"`
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	Direction         types.String `tfsdk:"direction"`
	Index             types.Int64  `tfsdk:"index"`
	Ports             []PortModel  `tfsdk:"ports"`
	Peers             []PeerModel  `tfsdk:"peers"`
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
		Description: "Manages a Security Rule within a Security Group in the ASC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the security rule (security_group_name:direction:index).",
				Computed:    true,
			},
			"security_group_name": schema.StringAttribute{
				Description: "The name of the parent security group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"direction": schema.StringAttribute{
				Description: "The rule direction: \"ingress\" or \"egress\".",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index": schema.Int64Attribute{
				Description: "The 0-based index of the rule within the direction (assigned by the API).",
				Computed:    true,
			},
			"ports": schema.ListNestedAttribute{
				Description: "Ports allowed by this rule.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Description: "The network protocol (TCP, UDP, or SCTP).",
							Required:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The port number (1-65535).",
							Required:    true,
						},
					},
				},
			},
			"peers": schema.ListNestedAttribute{
				Description: "Peers (sources for ingress, destinations for egress) for this rule.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pod_selector": schema.MapAttribute{
							Description: "Selects pods by label within the same namespace.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"namespace_selector": schema.MapAttribute{
							Description: "Selects namespaces by label.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"ip_block": schema.SingleNestedAttribute{
							Description: "Selects a CIDR range and optional exceptions.",
							Optional:    true,
							Attributes: map[string]schema.Attribute{
								"cidr": schema.StringAttribute{
									Description: "The IP range in CIDR notation (e.g. \"10.0.0.0/24\").",
									Required:    true,
								},
								"except": schema.ListAttribute{
									Description: "CIDRs to exclude from the range.",
									Optional:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
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

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected resource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = dataClient.Network
}

// Create creates a new security rule.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sgName := plan.SecurityGroupName.ValueString()
	direction := plan.Direction.ValueString()

	// Build the add-rule request entry
	entry := client.AddRuleEntry{
		Direction: direction,
	}

	// Convert ports
	for _, p := range plan.Ports {
		entry.Ports = append(entry.Ports, client.AddRulePortEntry{
			Protocol: p.Protocol.ValueString(),
			Port:     int(p.Port.ValueInt64()),
		})
	}

	// Convert peers
	for _, peer := range plan.Peers {
		peerEntry := client.AddRulePeerEntry{}

		if !peer.PodSelector.IsNull() {
			podSel := make(map[string]string)
			resp.Diagnostics.Append(peer.PodSelector.ElementsAs(ctx, &podSel, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			peerEntry.PodSelector = podSel
		}

		if !peer.NamespaceSelector.IsNull() {
			nsSel := make(map[string]string)
			resp.Diagnostics.Append(peer.NamespaceSelector.ElementsAs(ctx, &nsSel, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			peerEntry.NamespaceSelector = nsSel
		}

		if peer.IPBlock != nil {
			ipBlock := &client.AddRuleIPBlockEntry{
				CIDR: peer.IPBlock.CIDR.ValueString(),
			}
			if !peer.IPBlock.Except.IsNull() {
				var except []string
				resp.Diagnostics.Append(peer.IPBlock.Except.ElementsAs(ctx, &except, false)...)
				if resp.Diagnostics.HasError() {
					return
				}
				ipBlock.Except = except
			}
			peerEntry.IPBlock = ipBlock
		}

		entry.Peers = append(entry.Peers, peerEntry)
	}

	addReq := client.AddRulesRequest{Rules: []client.AddRuleEntry{entry}}

	sg, err := r.client.AddSecurityRules(ctx, sgName, addReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Security Rule",
			fmt.Sprintf("Could not add rule to security group %q: %s", sgName, err.Error()),
		)
		return
	}

	// Determine the index of the newly created rule (last rule in the direction)
	var ruleIndex int
	if direction == "ingress" {
		ruleIndex = len(sg.IngressRules) - 1
	} else {
		ruleIndex = len(sg.EgressRules) - 1
	}

	plan.Index = types.Int64Value(int64(ruleIndex))
	plan.ID = types.StringValue(createRuleStateID(sgName, direction, ruleIndex))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read reads the security rule from the API.
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
			fmt.Sprintf("Could not get security group %q: %s", sgName, err.Error()),
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
		resp.State.RemoveResource(ctx)
		return
	}

	rule := rules[index]

	// Update state from API response
	state.ID = types.StringValue(createRuleStateID(sgName, direction, index))
	state.Index = types.Int64Value(int64(index))

	state.Ports = make([]PortModel, len(rule.Ports))
	for i, p := range rule.Ports {
		state.Ports[i] = PortModel{
			Protocol: types.StringValue(p.Protocol),
			Port:     types.Int64Value(int64(p.Port)),
		}
	}

	state.Peers = make([]PeerModel, len(rule.Peers))
	for i, peer := range rule.Peers {
		pm := PeerModel{}

		if len(peer.PodSelector) > 0 {
			mapVal, diags := types.MapValueFrom(ctx, types.StringType, peer.PodSelector)
			resp.Diagnostics.Append(diags...)
			pm.PodSelector = mapVal
		} else {
			pm.PodSelector = types.MapNull(types.StringType)
		}

		if len(peer.NamespaceSelector) > 0 {
			mapVal, diags := types.MapValueFrom(ctx, types.StringType, peer.NamespaceSelector)
			resp.Diagnostics.Append(diags...)
			pm.NamespaceSelector = mapVal
		} else {
			pm.NamespaceSelector = types.MapNull(types.StringType)
		}

		if peer.IPBlock != nil {
			ipBlock := &PeerIPBlockModel{
				CIDR: types.StringValue(peer.IPBlock.CIDR),
			}
			if len(peer.IPBlock.Except) > 0 {
				exceptVal, diags := types.ListValueFrom(ctx, types.StringType, peer.IPBlock.Except)
				resp.Diagnostics.Append(diags...)
				ipBlock.Except = exceptVal
			} else {
				ipBlock.Except = types.ListNull(types.StringType)
			}
			pm.IPBlock = ipBlock
		}

		state.Peers[i] = pm
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported for security rules.
func (r *Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Security Rule updates are not supported. Delete and recreate the rule instead.",
	)
}

// Delete deletes the security rule.
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
			fmt.Sprintf("Could not delete security rule: %s", err.Error()),
		)
		return
	}
}

// createRuleStateID creates a unique identifier for the security rule resource.
func createRuleStateID(sgName, direction string, index int) string {
	return sgName + ":" + direction + ":" + strconv.Itoa(index)
}

// isNotFoundError checks if the error indicates a resource was not found.
func isNotFoundError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not found") ||
		len(err.Error()) > 14 && err.Error()[:14] == "API error 404:")
}
