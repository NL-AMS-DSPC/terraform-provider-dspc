package client

import (
	"context"
	"fmt"
	"net/http"
)

// ListenerRequest describes a single L4 frontend (protocol + port).
type ListenerRequest struct {
	Protocol string `json:"protocol"`
	Port     uint32 `json:"port"`
}

// BackendRequest describes a single backend target.
type BackendRequest struct {
	IP   string `json:"ip"`
	Port uint32 `json:"port"`
}

// HealthCheckRequest describes the backend health-check configuration.
type HealthCheckRequest struct {
	Protocol           string `json:"protocol"`
	Port               uint32 `json:"port"`
	IntervalSeconds    uint32 `json:"intervalSeconds"`
	TimeoutSeconds     uint32 `json:"timeoutSeconds"`
	HealthyThreshold   uint32 `json:"healthyThreshold"`
	UnhealthyThreshold uint32 `json:"unhealthyThreshold"`
}

// CreateLoadBalancerRequest represents the API request body for creating an L4 load balancer.
type CreateLoadBalancerRequest struct {
	Name        string              `json:"name"`
	Scheme      string              `json:"scheme"`
	SKUID       string              `json:"skuID"`
	VPCID       string              `json:"vpcID"`
	SubnetID    string              `json:"subnetID"`
	Listeners   []ListenerRequest   `json:"listeners"`
	Backends    []BackendRequest    `json:"backends"`
	Algorithm   string              `json:"algorithm"`
	HealthCheck *HealthCheckRequest `json:"healthCheck,omitempty"`
	Tags        []Tag               `json:"tags,omitempty"`
}

// CreateLoadBalancerResponse represents the API response after creating an L4 load balancer.
type CreateLoadBalancerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type loadBalancerClient struct {
	apiClient
}

// CreateLoadBalancer creates a new load balancer.
func (api *loadBalancerClient) CreateLoadBalancer(ctx context.Context, createRequest CreateLoadBalancerRequest) (lb *CreateLoadBalancerResponse, err error) {
	err = api.post(ctx, api.namespacedPath("/load-balancers"), createRequest, &lb)
	return
}

// DeleteLoadBalancer deletes a load balancer by name.
func (api *loadBalancerClient) DeleteLoadBalancer(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/load-balancers/%s", name)))
}

func newLoadBalancerClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *loadBalancerClient {
	return &loadBalancerClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
