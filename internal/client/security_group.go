package client

import (
	"context"
	"fmt"
)

// SecurityGroup represents a security group in the ASC network API
type SecurityGroup struct {
	Name         string         `json:"name"`
	Namespace    string         `json:"namespace,omitempty"`
	IngressRules []SecurityRule `json:"ingressRules,omitempty"`
	EgressRules  []SecurityRule `json:"egressRules,omitempty"`
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

// AddRuleEntry represents a single rule entry in the add-rules request
type AddRuleEntry struct {
	Direction string             `json:"direction"`
	Peers     []AddRulePeerEntry `json:"peers,omitempty"`
	Ports     []AddRulePortEntry `json:"ports,omitempty"`
}

// AddRulePeerEntry represents a peer in an add-rule request
type AddRulePeerEntry struct {
	PodSelector       map[string]string    `json:"podSelector,omitempty"`
	NamespaceSelector map[string]string    `json:"namespaceSelector,omitempty"`
	IPBlock           *AddRuleIPBlockEntry `json:"ipBlock,omitempty"`
}

// AddRuleIPBlockEntry represents an IP block in an add-rule request
type AddRuleIPBlockEntry struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// AddRulePortEntry represents a port in an add-rule request
type AddRulePortEntry struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

// AddRulesRequest represents the HTTP request body for adding rules to a security group
type AddRulesRequest struct {
	Rules []AddRuleEntry `json:"rules"`
}

// ListRulesResponse is the HTTP response body for listing rules of a Security Group
type ListRulesResponse struct {
	Ingress []SecurityRule `json:"ingress"`
	Egress  []SecurityRule `json:"egress"`
}

// CreateSecurityGroup creates a new security group
func (api *networkClient) CreateSecurityGroup(ctx context.Context, name string) (sg *SecurityGroup, err error) {
	err = api.post(ctx, api.namespacedPath("/security-groups"), CreateSecurityGroupRequest{Name: name}, &sg)
	return
}

// GetSecurityGroup retrieves a security group by name
func (api *networkClient) GetSecurityGroup(ctx context.Context, name string) (sg *SecurityGroup, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s", name)), &sg)
	return
}

// ListSecurityGroups retrieves all security groups
func (api *networkClient) ListSecurityGroups(ctx context.Context) (sgs []*SecurityGroup, err error) {
	err = api.get(ctx, api.namespacedPath("/security-groups"), &sgs)
	return
}

// DeleteSecurityGroup deletes a security group by name
func (api *networkClient) DeleteSecurityGroup(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s", name)))
}

// AddSecurityRules adds one or more rules to a security group
func (api *networkClient) AddSecurityRules(ctx context.Context, sgName string, req AddRulesRequest) (sg *SecurityGroup, err error) {
	err = api.post(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/rules", sgName)), req, &sg)
	return
}

// ListSecurityRules retrieves all rules for a security group
func (api *networkClient) ListSecurityRules(ctx context.Context, sgName string) (resp *ListRulesResponse, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/rules", sgName)), &resp)
	return
}

// DeleteSecurityRule deletes a specific rule by direction and index
func (api *networkClient) DeleteSecurityRule(ctx context.Context, sgName, direction string, index int) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/rules/%s/%d", sgName, direction, index)))
}

// SecurityGroupAttachment represents an attachment between a Security Group and a target resource
type SecurityGroupAttachment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	SGRef     string `json:"sgRef"`
}

// AttachSecurityGroupRequest represents the request body for attaching a security group
type AttachSecurityGroupRequest struct {
	TargetType string `json:"targetType"`
	TargetName string `json:"targetName"`
}

// AttachSecurityGroup attaches a security group to a target resource
func (api *networkClient) AttachSecurityGroup(ctx context.Context, sgName, targetType, targetName string) (*SecurityGroupAttachment, error) {
	var sga *SecurityGroupAttachment
	err := api.post(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/attach", sgName)), AttachSecurityGroupRequest{
		TargetType: targetType,
		TargetName: targetName,
	}, &sga)
	return sga, err
}

// GetSecurityGroupAttachment retrieves a specific security group attachment
func (api *networkClient) GetSecurityGroupAttachment(ctx context.Context, sgName, attachmentName string) (*SecurityGroupAttachment, error) {
	var sga *SecurityGroupAttachment
	err := api.get(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/attachments/%s", sgName, attachmentName)), &sga)
	return sga, err
}

// ListSecurityGroupAttachments lists all attachments for a security group
func (api *networkClient) ListSecurityGroupAttachments(ctx context.Context, sgName string) ([]*SecurityGroupAttachment, error) {
	var sgas []*SecurityGroupAttachment
	err := api.get(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/attachments", sgName)), &sgas)
	return sgas, err
}

// DetachSecurityGroup detaches a security group from a target resource
func (api *networkClient) DetachSecurityGroup(ctx context.Context, sgName, attachmentName string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/security-groups/%s/attachments/%s", sgName, attachmentName)))
}
