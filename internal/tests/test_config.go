package tests

// BuildTestPath constructs a test API path using the default service configuration
// These paths match the DefaultServiceConfig in internal/client/config.go
func BuildTestPath(service, resourcePath string) string {
	var prefix string
	switch service {
	case "vm":
		prefix = "/api/vm"
	case "network":
		prefix = "/api/network"
	case "storage":
		prefix = "/api/vm" // BlockStorage shares path with VM service
	default:
		prefix = service // fallback for custom paths
	}
	
	return prefix + "/v1/namespaces/test-ns" + resourcePath
}
