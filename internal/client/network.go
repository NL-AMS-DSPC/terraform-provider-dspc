package client

import (
	"context"
	"fmt"
	"net/http"
)

// VPC represents a Virtual Private Cloud in the DSPC network API
type VPC struct {
	ID        string   `json:"id"`
	URN       string   `json:"urn"`
	Name      string   `json:"name"`
	CIDR      string   `json:"cidr"`
	Status    string   `json:"status"`
	LastError string   `json:"lastError" `
	Subnets   []Subnet `json:"subnets"`
	Tags      []Tag    `json:"tags"`
}

// CreateVPCRequest represents the request body for creating a VPC
type CreateVPCRequest struct {
	ID      string                `json:"id,omitempty"`
	Name    string                `json:"name"`
	Tags    []Tag                 `json:"tags,omitempty"`
	Subnets []CreateSubnetRequest `json:"subnets,omitempty"`
}

// CreateVPCResponse represents the API response after creating a VPC.
type CreateVPCResponse struct {
	ID   string `json:"id"`
	URN  string `json:"urn"`
	Name string `json:"name"`
}

// Subnet represents a subnet within a VPC in the DSPC network API
type Subnet struct {
	ID        string `json:"id"`
	URN       string `json:"urn"`
	Name      string `json:"name"`
	CIDR      string `json:"cidr"`
	Type      string `json:"type"`
	VPCID     string `json:"vpcID"`
	Status    string `json:"status"`
	LastError string `json:"lastError"`
	Tags      []Tag  `json:"tags"`
}

// CreateSubnetRequest represents the request body for creating a subnet
type CreateSubnetRequest struct {
	Name  string `json:"name"`
	CIDR  string `json:"cidr"`
	VPCID string `json:"vpcID"`
	Type  string `json:"type"`
	Tags  []Tag  `json:"tags,omitempty"`
}

// CreateSubnetResponse represents the API response after creating a Subnet.
type CreateSubnetResponse struct {
	ID      string `json:"id"`
	URN     string `json:"urn"`
	Created string `json:"created"`
}

type networkClient struct {
	apiClient
}

// CreateVPC creates a new VPC
func (api *networkClient) CreateVPC(ctx context.Context, id, name string, tags []Tag, subnets []CreateSubnetRequest) (resp CreateVPCResponse, err error) {
	err = api.post(ctx, api.namespacedPath("/vpcs"), CreateVPCRequest{
		ID:      id,
		Name:    name,
		Tags:    tags,
		Subnets: subnets,
	}, &resp)
	return
}

// GetVPC retrieves a VPC by name
func (api *networkClient) GetVPC(ctx context.Context, name string) (vpc VPC, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s", name)), &vpc)
	return
}

// ListVPCs retrieves all VPCs
func (api *networkClient) ListVPCs(ctx context.Context) (vpcs []VPC, err error) {
	err = api.get(ctx, api.namespacedPath("/vpcs"), &vpcs)
	return
}

// DeleteVPC deletes a VPC by name
func (api *networkClient) DeleteVPC(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s", name)))
}

// CreateSubnet creates a new subnet within a VPC
func (api *networkClient) CreateSubnet(ctx context.Context, vpcName, vpcID, name, cidr, subnetType string, tags []Tag) (resp CreateSubnetResponse, err error) {
	err = api.post(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s/subnets", vpcName)), CreateSubnetRequest{
		Name:  name,
		VPCID: vpcID,
		CIDR:  cidr,
		Type:  subnetType,
		Tags:  tags,
	}, &resp)
	return
}

// ListSubnetsForVPC retrieves all subnets for a VPC
func (api *networkClient) ListSubnetsForVPC(ctx context.Context, vpcName string) (subnets []Subnet, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s/subnets", vpcName)), &subnets)
	return
}

// DeleteSubnet deletes a subnet within a VPC
func (api *networkClient) DeleteSubnet(ctx context.Context, vpcName, subnetName string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s/subnets/%s", vpcName, subnetName)))
}

func newNetworkClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *networkClient {
	return &networkClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
