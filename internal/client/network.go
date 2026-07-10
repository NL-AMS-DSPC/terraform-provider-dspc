package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// VPC represents a Virtual Private Cloud in the DSPC network API
type VPC struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name"`
	CIDR            string   `json:"cidr"`
	Status          string   `json:"status"`
	Tags            []Tag    `json:"tags,omitempty"`
	Subnets         []Subnet `json:"subnets,omitempty"`
	ResourceVersion string   `json:"resourceVersion,omitempty"`
	Namespace       string   `json:"namespace,omitempty"`
}

// CreateVPCRequest represents the request body for creating a VPC
type CreateVPCRequest struct {
	ID      string   `json:"id,omitempty"`
	Name    string   `json:"name"`
	Tags    []Tag    `json:"tags,omitempty"`
	Subnets []Subnet `json:"subnets,omitempty"`
}

// Subnet represents a subnet within a VPC in the DSPC network API
type Subnet struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	TenantID string    `json:"tenantID"`
	CIDR     string    `json:"cidr"`
	Type     string    `json:"type"`
	VPCID    uuid.UUID `json:"vpcID"`
	Status   string    `json:"status,omitempty"`
	Tags     []Tag     `json:"tags,omitempty"`
}

// CreateSubnetRequest represents the request body for creating a subnet
type CreateSubnetRequest struct {
	Name  string `json:"name" `
	CIDR  string `json:"cidr"`
	VPCID string `json:"vpcID"`
	Type  string `json:"type"`
	Tags  []Tag  `json:"tags,omitempty"`
}

type CreateSubnetResponse struct {
	ID      string `json:"id"`
	URN     string `json:"urn"`
	Created string `json:"created"`
}

type networkClient struct {
	apiClient
}

// CreateVPC creates a new VPC
func (api *networkClient) CreateVPC(ctx context.Context, id, name string, tags []Tag, subnets []Subnet) (vpc *VPC, err error) {
	err = api.post(ctx, api.namespacedPath("/vpcs"), CreateVPCRequest{
		ID:      id,
		Name:    name,
		Tags:    tags,
		Subnets: subnets,
	}, &vpc)
	return
}

// GetVPC retrieves a VPC by name
func (api *networkClient) GetVPC(ctx context.Context, name string) (vpc *VPC, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s", name)), &vpc)
	return
}

// ListVPCs retrieves all VPCs
func (api *networkClient) ListVPCs(ctx context.Context) (vpcs []*VPC, err error) {
	err = api.get(ctx, api.namespacedPath("/vpcs"), &vpcs)
	return
}

// DeleteVPC deletes a VPC by name
func (api *networkClient) DeleteVPC(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s", name)))
}

// CreateSubnet creates a new subnet within a VPC
func (api *networkClient) CreateSubnet(ctx context.Context, vpcName, name, cidr, vpcID, subnetType string, tags []Tag) (subnet *CreateSubnetResponse, err error) {
	err = api.post(ctx, api.namespacedPath(fmt.Sprintf("/vpcs/%s/subnets", vpcName)), CreateSubnetRequest{
		Name:  name,
		CIDR:  cidr,
		VPCID: vpcID,
		Type:  subnetType,
		Tags:  tags,
	}, &subnet)
	return
}

// ListSubnetsForVPC retrieves all subnets for a VPC
func (api *networkClient) ListSubnetsForVPC(ctx context.Context, vpcName string) (subnets []*Subnet, err error) {
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
