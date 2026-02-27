// Package securityrule provides Terraform resources and data sources for managing Security Rules.
package securityrule

import (
	"context"
	"fmt"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &DataSource{}
	_ datasource.DataSourceWithConfigure = &DataSource{}
)

// DataClient defines the interface for listing security rules.
type DataClient interface {
	ListSecurityRules(ctx context.Context, sgName string) (*client.ListRulesResponse, error)
}

// DataSource defines the data source implementation.
type DataSource struct {
	client DataClient
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	SecurityGroupName types.String    `tfsdk:"security_group_name"`
	IngressRules      []RuleModel    `tfsdk:"ingress_rules"`
	EgressRules       []RuleModel    `tfsdk:"egress_rules"`
}

// RuleModel describes a single security rule in the data source.
type RuleModel struct {
	Index types.Int64        `tfsdk:"index"`
	Peers []DataSourcePeer   `tfsdk:"peers"`
	Ports []DataSourcePort   `tfsdk:"ports"`
}

// DataSourcePeer represents a security peer in the data source schema.
type DataSourcePeer struct {
	PodSelector       types.Map    `tfsdk:"pod_selector"`
	NamespaceSelector types.Map    `tfsdk:"namespace_selector"`
	IPBlockCIDR       types.String `tfsdk:"ip_block_cidr"`
	IPBlockExcept     []types.String `tfsdk:"ip_block_except"`
}

// DataSourcePort represents a security port in the data source schema.
type DataSourcePort struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int64  `tfsdk:"port"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_rules"
}

func ruleNestedSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"index": schema.Int64Attribute{
				Description: "The 0-based position of this rule.",
				Computed:    true,
			},
			"peers": schema.ListNestedAttribute{
				Description: "The peers for this rule.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pod_selector": schema.MapAttribute{
							Description: "Selects pods by label within the same namespace.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"namespace_selector": schema.MapAttribute{
							Description: "Selects namespaces by label.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"ip_block_cidr": schema.StringAttribute{
							Description: "IP range in CIDR notation.",
							Computed:    true,
						},
						"ip_block_except": schema.ListAttribute{
							Description: "CIDRs to exclude from the IP block range.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"ports": schema.ListNestedAttribute{
				Description: "The ports allowed by this rule.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"protocol": schema.StringAttribute{
							Description: "The network protocol (TCP, UDP, or SCTP).",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The port number.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Schema updates the data source schema with the attributes.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all ingress and egress rules for a Security Group from the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"security_group_name": schema.StringAttribute{
				Description: "The name of the Security Group to list rules for.",
				Required:    true,
			},
			"ingress_rules": schema.ListNestedAttribute{
				Description: "The ingress (inbound) rules.",
				Computed:    true,
				NestedObject: ruleNestedSchema(),
			},
			"egress_rules": schema.ListNestedAttribute{
				Description: "The egress (outbound) rules.",
				Computed:    true,
				NestedObject: ruleNestedSchema(),
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected data source configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Network
}

// Read fetches the list of rules for a Security Group and stores them in the state.
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sgName := config.SecurityGroupName.ValueString()

	rulesResp, err := d.client.ListSecurityRules(ctx, sgName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing Security Rules",
			fmt.Sprintf("Could not list rules for Security Group %q: %s", sgName, err.Error()),
		)
		return
	}

	config.IngressRules = flattenRulesForDataSource(rulesResp.Ingress)
	config.EgressRules = flattenRulesForDataSource(rulesResp.Egress)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// flattenRulesForDataSource converts a slice of client.SecurityRule to the data source model.
func flattenRulesForDataSource(rules []client.SecurityRule) []RuleModel {
	result := make([]RuleModel, len(rules))
	for i, rule := range rules {
		result[i] = RuleModel{
			Index: types.Int64Value(int64(rule.Index)),
			Peers: flattenPeersForDataSource(rule.Peers),
			Ports: flattenPortsForDataSource(rule.Ports),
		}
	}
	return result
}

// flattenPeersForDataSource converts client.SecurityPeer slice to data source peers.
func flattenPeersForDataSource(peers []client.SecurityPeer) []DataSourcePeer {
	result := make([]DataSourcePeer, len(peers))
	for i, p := range peers {
		peer := DataSourcePeer{
			PodSelector:       types.MapNull(types.StringType),
			NamespaceSelector: types.MapNull(types.StringType),
			IPBlockCIDR:       types.StringNull(),
			IPBlockExcept:     nil,
		}

		if len(p.PodSelector) > 0 {
			peer.PodSelector = stringMapToTerraform(p.PodSelector)
		}

		if len(p.NamespaceSelector) > 0 {
			peer.NamespaceSelector = stringMapToTerraform(p.NamespaceSelector)
		}

		if p.IPBlock != nil {
			peer.IPBlockCIDR = types.StringValue(p.IPBlock.CIDR)
			if len(p.IPBlock.Except) > 0 {
				excepts := make([]types.String, len(p.IPBlock.Except))
				for j, e := range p.IPBlock.Except {
					excepts[j] = types.StringValue(e)
				}
				peer.IPBlockExcept = excepts
			}
		}

		result[i] = peer
	}
	return result
}

// flattenPortsForDataSource converts client.SecurityPort slice to data source ports.
func flattenPortsForDataSource(ports []client.SecurityPort) []DataSourcePort {
	result := make([]DataSourcePort, len(ports))
	for i, p := range ports {
		result[i] = DataSourcePort{
			Protocol: types.StringValue(p.Protocol),
			Port:     types.Int64Value(int64(p.Port)),
		}
	}
	return result
}

// stringMapToTerraform converts a Go string map to a Terraform MapValue.
func stringMapToTerraform(m map[string]string) types.Map {
	if len(m) == 0 {
		return types.MapNull(types.StringType)
	}

	attrElems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		attrElems[k] = types.StringValue(v)
	}

	result, _ := types.MapValue(types.StringType, attrElems)
	return result
}
