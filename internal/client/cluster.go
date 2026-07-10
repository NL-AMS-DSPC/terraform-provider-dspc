package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ClusterNodePoolRequest is a node pool replica + SKU specification submitted on cluster creation.
type ClusterNodePoolRequest struct {
	Replicas int32  `json:"replicas"`
	SKUID    string `json:"skuId"`
}

// ClusterSubnetRequest identifies a subnet by name when linking a VPC to the cluster.
type ClusterSubnetRequest struct {
	Name string `json:"name"`
}

// ClusterSubnetsRequest groups the pod and service subnet names.
type ClusterSubnetsRequest struct {
	Pods     ClusterSubnetRequest `json:"pods"`
	Services ClusterSubnetRequest `json:"services"`
}

// ClusterVPCRequest specifies the VPC and its pod/service subnets to attach the cluster to.
type ClusterVPCRequest struct {
	Name    string                `json:"name"`
	Subnets ClusterSubnetsRequest `json:"subnets"`
}

// ClusterCreateRequest is the create-cluster payload.
// PullSecret and SSHKey are write-only and never returned by the API on read.
type ClusterCreateRequest struct {
	Name         string                 `json:"name"`
	Domain       string                 `json:"domain"`
	Version      string                 `json:"version"`
	Image        string                 `json:"image"`
	Tags         []Tag                  `json:"tags,omitempty"`
	ControlPlane ClusterNodePoolRequest `json:"controlPlane"`
	Workers      ClusterNodePoolRequest `json:"workers"`
	VPC          ClusterVPCRequest      `json:"vpc"`
	PullSecret   string                 `json:"pullSecret"`
	SSHKey       string                 `json:"sshKey"`
}

// ClusterPatchRequest is the patch-cluster payload (tags only).
// Tags has no omitempty so an empty slice marshals as "tags": []
// and the API receives an explicit signal to clear all tags.
type ClusterPatchRequest struct {
	Tags []Tag `json:"tags"`
}

// ClusterNode is one node's status as returned in a cluster node pool response.
type ClusterNode struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ClusterNodePool is the node-pool projection on a cluster read response.
type ClusterNodePool struct {
	Replicas int32         `json:"replicas"`
	SKUID    string        `json:"skuId"`
	Nodes    []ClusterNode `json:"nodes,omitempty"`
}

// ClusterSubnet is a subnet's name and CIDR on a cluster read response.
type ClusterSubnet struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
}

// ClusterSubnets groups the pod and service subnet projections on read.
type ClusterSubnets struct {
	Pods     ClusterSubnet `json:"pods"`
	Services ClusterSubnet `json:"services"`
}

// ClusterVPC is the VPC projection on a cluster read response.
type ClusterVPC struct {
	Name    string         `json:"name"`
	CIDR    string         `json:"cidr"`
	Subnets ClusterSubnets `json:"subnets"`
}

// Cluster is the read projection returned by cluster-service.
type Cluster struct {
	URN           string          `json:"urn"`
	Name          string          `json:"name"`
	Domain        string          `json:"domain"`
	Version       string          `json:"version"`
	Image         string          `json:"image"`
	Status        string          `json:"status"`
	StatusMessage string          `json:"statusMessage,omitempty"`
	Tags          []Tag           `json:"tags,omitempty"`
	ControlPlane  ClusterNodePool `json:"controlPlane"`
	Workers       ClusterNodePool `json:"workers"`
	VPC           ClusterVPC      `json:"vpc"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// ClusterNodeStatus is the per-node entry inside a cluster status report.
type ClusterNodeStatus struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	VMStatus string `json:"vmStatus"`
	Healthy  bool   `json:"healthy"`
}

// ClusterStatusReport is the rolled-up health + per-node status of a cluster.
type ClusterStatusReport struct {
	Name          string              `json:"name"`
	Status        string              `json:"status"`
	StatusMessage string              `json:"statusMessage,omitempty"`
	Healthy       bool                `json:"healthy"`
	NodesTotal    int                 `json:"nodesTotal"`
	NodesReady    int                 `json:"nodesReady"`
	Nodes         []ClusterNodeStatus `json:"nodes,omitempty"`
}

type clusterClient struct {
	apiClient
}

// CreateCluster creates a new cluster.
// cluster-service returns 202 with the persisted cluster; provisioning continues asynchronously.
func (api *clusterClient) CreateCluster(ctx context.Context, req ClusterCreateRequest) (*Cluster, error) {
	var out Cluster
	if err := api.post(ctx, "/v1/clusters", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListClusters returns all clusters for the authenticated tenant.
func (api *clusterClient) ListClusters(ctx context.Context) ([]Cluster, error) {
	var out []Cluster
	if err := api.get(ctx, "/v1/clusters", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCluster fetches a cluster by its name.
func (api *clusterClient) GetCluster(ctx context.Context, name string) (*Cluster, error) {
	var out Cluster
	if err := api.get(ctx, fmt.Sprintf("/v1/clusters/%s", name), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetClusterStatus returns the rolled-up lifecycle + per-node status for a cluster.
func (api *clusterClient) GetClusterStatus(ctx context.Context, name string) (*ClusterStatusReport, error) {
	var out ClusterStatusReport
	if err := api.get(ctx, fmt.Sprintf("/v1/clusters/%s/status", name), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PatchCluster updates a cluster's tags. Pass an empty slice to clear all tags.
func (api *clusterClient) PatchCluster(ctx context.Context, name string, tags []Tag) (*Cluster, error) {
	var out Cluster
	body := ClusterPatchRequest{Tags: tags}
	if err := api.patch(ctx, fmt.Sprintf("/v1/clusters/%s", name), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCluster removes a cluster and its associated VMs.
func (api *clusterClient) DeleteCluster(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/clusters/%s", name))
}

func newClusterClient(endpoint, pathPrefix string, authMgr *authManager, httpClient *http.Client) *clusterClient {
	return &clusterClient{
		newAPIClient(endpoint, "", pathPrefix, authMgr, httpClient),
	}
}
