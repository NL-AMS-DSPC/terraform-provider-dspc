package client

import (
	"context"
	"fmt"
	"net/http"
)

// VPC represents a Virtual Private Cloud in the DSPC network API
type VPC struct {
	Name            string   `json:"name"`
	CIDR            string   `json:"cidr"`
	Status          string   `json:"status"`
	Subnets         []Subnet `json:"subnets,omitempty"`
	ResourceVersion string   `json:"resourceVersion,omitempty"`
	Namespace       string   `json:"namespace,omitempty"`
}

// CreateVPCRequest represents the request body for creating a VPC
type CreateVPCRequest struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
}

// Subnet represents a subnet within a VPC in the DSPC network API
type Subnet struct {
	Name            string `json:"name"`
	CIDR            string `json:"cidr"`
	Type            string `json:"type"`
	VPCRef          string `json:"vpcRef"`
	Status          string `json:"status,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// CreateSubnetRequest represents the request body for creating a subnet
type CreateSubnetRequest struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
	Type string `json:"type"`
}

// SecurityGroup represents a network security group in the DSPC network API
type SecurityGroup struct {
	Name            string         `json:"name"`
	Namespace       string         `json:"namespace,omitempty"`
	IngressRules    []SecurityRule `json:"ingressRules,omitempty"`
	EgressRules     []SecurityRule `json:"egressRules,omitempty"`
	ResourceVersion string         `json:"resourceVersion,omitempty"`
}

// SecurityRule represents a single ingress or egress traffic rule
type SecurityRule struct {
	Index int            `json:"index"`
	Peers []SecurityPeer `json:"peers,omitempty"`
	Ports []SecurityPort `json:"ports,omitempty"`
}

// SecurityPeer identifies a set of pods or IP ranges that traffic is allowed to/from
type SecurityPeer struct {
	PodSelector       map[string]string `json:"podSelector,omitempty"`
	NamespaceSelector map[string]string `json:"namespaceSelector,omitempty"`
	IPBlock           *IPBlock          `json:"ipBlock,omitempty"`
}

// IPBlock describes a particular CIDR range with optional exception CIDRs
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// SecurityPort defines a protocol and port number allowed by a rule
type SecurityPort struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

// CreateSecurityGroupRequest represents the request body for creating a security group
type CreateSecurityGroupRequest struct {
	Name string `json:"name"`
}

// AddRuleRequest represents a single rule entry for adding a rule to a security group
type AddRuleRequest struct {
	Direction string       `json:"direction"`
	Rule      SecurityRule `json:"rule"`
}

// AddRulesRequest represents the request body for adding rules to a security group
type AddRulesRequest struct {
	Rules []AddRuleRequest `json:"rules"`
}

// ListRulesResponse represents the response for listing rules of a security group
type ListRulesResponse struct {
	Ingress []SecurityRule `json:"ingress"`
	Egress  []SecurityRule `json:"egress"`
}

type networkClient struct {
	apiClient
}

// CreateVPC creates a new VPC
func (api *networkClient) CreateVPC(ctx context.Context, name, cidr string) (vpc *VPC, err error) {
	err = api.post(ctx, "/vpcs", CreateVPCRequest{Name: name, CIDR: cidr}, &vpc)
	return
}

// GetVPC retrieves a VPC by name
func (api *networkClient) GetVPC(ctx context.Context, name string) (vpc *VPC, err error) {
	err = api.get(ctx, fmt.Sprintf("/vpcs/%s", name), &vpc)
	return
}

// ListVPCs retrieves all VPCs
func (api *networkClient) ListVPCs(ctx context.Context) (vpcs []*VPC, err error) {
	err = api.get(ctx, "/vpcs", &vpcs)
	return
}

// DeleteVPC deletes a VPC by name
func (api *networkClient) DeleteVPC(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/vpcs/%s", name))
}

// CreateSubnet creates a new subnet within a VPC
func (api *networkClient) CreateSubnet(ctx context.Context, vpcName, name, cidr, subnetType string) (subnet *Subnet, err error) {
	err = api.post(ctx, fmt.Sprintf("/vpcs/%s/subnets", vpcName), CreateSubnetRequest{
		Name: name,
		CIDR: cidr,
		Type: subnetType,
	}, &subnet)
	return
}

// ListSubnetsForVPC retrieves all subnets for a VPC
func (api *networkClient) ListSubnetsForVPC(ctx context.Context, vpcName string) (subnets []*Subnet, err error) {
	err = api.get(ctx, fmt.Sprintf("/vpcs/%s/subnets", vpcName), &subnets)
	return
}

// DeleteSubnet deletes a subnet within a VPC
func (api *networkClient) DeleteSubnet(ctx context.Context, vpcName, subnetName string) error {
	return api.delete(ctx, fmt.Sprintf("/vpcs/%s/subnets/%s", vpcName, subnetName))
}

// CreateSecurityGroup creates a new security group
func (api *networkClient) CreateSecurityGroup(ctx context.Context, name string) (sg *SecurityGroup, err error) {
	err = api.post(ctx, "/security-groups", CreateSecurityGroupRequest{Name: name}, &sg)
	return
}

// GetSecurityGroup retrieves a security group by name
func (api *networkClient) GetSecurityGroup(ctx context.Context, name string) (sg *SecurityGroup, err error) {
	err = api.get(ctx, fmt.Sprintf("/security-groups/%s", name), &sg)
	return
}

// ListSecurityGroups retrieves all security groups
func (api *networkClient) ListSecurityGroups(ctx context.Context) (sgs []*SecurityGroup, err error) {
	err = api.get(ctx, "/security-groups", &sgs)
	return
}

// DeleteSecurityGroup deletes a security group by name
func (api *networkClient) DeleteSecurityGroup(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/security-groups/%s", name))
}

// AddSecurityRules adds one or more rules to a security group
func (api *networkClient) AddSecurityRules(ctx context.Context, sgName string, rules []AddRuleRequest) (sg *SecurityGroup, err error) {
	err = api.post(ctx, fmt.Sprintf("/security-groups/%s/rules", sgName), AddRulesRequest{Rules: rules}, &sg)
	return
}

// ListSecurityRules retrieves all rules for a security group
func (api *networkClient) ListSecurityRules(ctx context.Context, sgName string) (resp *ListRulesResponse, err error) {
	err = api.get(ctx, fmt.Sprintf("/security-groups/%s/rules", sgName), &resp)
	return
}

// DeleteSecurityRule deletes a specific rule by direction and index
func (api *networkClient) DeleteSecurityRule(ctx context.Context, sgName, direction string, index int) error {
	return api.delete(ctx, fmt.Sprintf("/security-groups/%s/rules/%s/%d", sgName, direction, index))
}

func newNetworkClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *networkClient {
	return &networkClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
