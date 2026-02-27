package tests

import "github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"

// TestServiceConfig returns the service configuration used in tests
// This ensures test paths match the actual client configuration
func TestServiceConfig() client.ServiceConfig {
	return client.DefaultServiceConfig()
}

// BuildTestPath constructs a test API path using the service configuration
func BuildTestPath(service, resourcePath string) string {
	cfg := TestServiceConfig()
	
	var prefix string
	switch service {
	case "vm":
		prefix = cfg.VM.PathPrefix
	case "network":
		prefix = cfg.Network.PathPrefix
	case "storage":
		prefix = cfg.BlockStorage.PathPrefix
	default:
		prefix = service // fallback for custom paths
	}
	
	return prefix + "/v1/namespaces/test-ns" + resourcePath
}
