package securityrule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
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

// RuleModel represents a single security rule in the data source.
type RuleModel struct {
	Index    types.Int64       `tfsdk:"index"`
	Peers    []DataPeerModel   `tfsdk:"peers"`
	Ports    []DataPortModel   `tfsdk:"ports"`
}

// DataPortModel represents a port in the data source.
type DataPortModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Port     types.Int64  `tfsdk:"port"`
}

// DataPeerModel represents a peer in the data source.
type DataPeerModel struct {
	PodSelector       types.Map              `tfsdk:"pod_selector"`
	NamespaceSelector types.Map              `tfsdk:"namespace_selector"`
	IPBlock           *DataPeerIPBlockModel  `tfsdk:"ip_block"`
}

// DataPeerIPBlockModel represents an IP block in the data source.
type DataPeerIPBlockModel struct {
	CIDR   types.String `tfsdk:"cidr"`
	Except types.List   `tfsdk:"except"`
}

// DataSourceModel describes the data source data model.
type DataSourceModel struct {
	SecurityGroupName types.String `tfsdk:"security_group_name"`
	Ingress           []RuleModel  `tfsdk:"ingress"`
	Egress            []RuleModel  `tfsdk:"egress"`
}

// NewDataSource creates a new DataSource.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_rules"
}

// Schema updates the data source schema with the attributes.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	peerNestedAttrs := map[string]schema.Attribute{
		"pod_selector": schema.MapAttribute{
			Description: "Selects pods by label.",
			Computed:    true,
			ElementType: types.StringType,
		},
		"namespace_selector": schema.MapAttribute{
			Description: "Selects namespaces by label.",
			Computed:    true,
			ElementType: types.StringType,
		},
		"ip_block": schema.SingleNestedAttribute{
			Description: "IP block selector.",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"cidr": schema.StringAttribute{
					Description: "The IP range in CIDR notation.",
					Computed:    true,
				},
				"except": schema.ListAttribute{
					Description: "CIDRs to exclude.",
					Computed:    true,
					ElementType: types.StringType,
				},
			},
		},
	}

	portNestedAttrs := map[string]schema.Attribute{
		"protocol": schema.StringAttribute{
			Description: "The network protocol (TCP, UDP, or SCTP).",
			Computed:    true,
		},
		"port": schema.Int64Attribute{
			Description: "The port number.",
			Computed:    true,
		},
	}

	ruleNestedAttrs := map[string]schema.Attribute{
		"index": schema.Int64Attribute{
			Description: "The 0-based index of the rule.",
			Computed:    true,
		},
		"peers": schema.ListNestedAttribute{
			Description: "Peers for this rule.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: peerNestedAttrs,
			},
		},
		"ports": schema.ListNestedAttribute{
			Description: "Ports for this rule.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: portNestedAttrs,
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Retrieves ingress and egress rules for a Security Group in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"security_group_name": schema.StringAttribute{
				Description: "The name of the security group to list rules for.",
				Required:    true,
			},
			"ingress": schema.ListNestedAttribute{
				Description: "Ingress rules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: ruleNestedAttrs,
				},
			},
			"egress": schema.ListNestedAttribute{
				Description: "Egress rules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: ruleNestedAttrs,
				},
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
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	if dataClient.Network == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected network service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.Network
}

// Read fetches the rules for a security group and stores them in the state.
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
			fmt.Sprintf("Could not list rules for security group %q: %s", sgName, err.Error()),
		)
		return
	}

	config.Ingress = convertRules(ctx, rulesResp.Ingress, resp)
	config.Egress = convertRules(ctx, rulesResp.Egress, resp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func convertRules(ctx context.Context, rules []client.SecurityRule, resp *datasource.ReadResponse) []RuleModel {
	models := make([]RuleModel, len(rules))
	for i, rule := range rules {
		rm := RuleModel{
			Index: types.Int64Value(int64(rule.Index)),
		}

		rm.Ports = make([]DataPortModel, len(rule.Ports))
		for j, p := range rule.Ports {
			rm.Ports[j] = DataPortModel{
				Protocol: types.StringValue(p.Protocol),
				Port:     types.Int64Value(int64(p.Port)),
			}
		}

		rm.Peers = make([]DataPeerModel, len(rule.Peers))
		for j, peer := range rule.Peers {
			pm := DataPeerModel{}

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
				ipBlock := &DataPeerIPBlockModel{
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

			rm.Peers[j] = pm
		}

		models[i] = rm
	}
	return models
}
