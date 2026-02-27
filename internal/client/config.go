package client

import "os"

// ServiceEndpoint represents configuration for a service endpoint
type ServiceEndpoint struct {
	PathPrefix string
}

// ServiceConfig holds configuration for all DSPC API services
type ServiceConfig struct {
	VM      ServiceEndpoint
	Network ServiceEndpoint
	Storage ServiceEndpoint
}

// DefaultServiceConfig returns the default service configuration
// that matches the current DSPC API structure behind Envoy gateway
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		VM:      ServiceEndpoint{PathPrefix: "/api/vm"},
		Network: ServiceEndpoint{PathPrefix: "/api/network"},
		Storage: ServiceEndpoint{PathPrefix: "/api/vm"}, // Shares path with VM service
	}
}

// LoadServiceConfig loads service configuration with environment variable overrides
func LoadServiceConfig() ServiceConfig {
	cfg := DefaultServiceConfig()

	// Allow override via environment variables for different deployments
	if prefix := os.Getenv("DSPC_VM_PATH_PREFIX"); prefix != "" {
		cfg.VM.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_NETWORK_PATH_PREFIX"); prefix != "" {
		cfg.Network.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_STORAGE_PATH_PREFIX"); prefix != "" {
		cfg.Storage.PathPrefix = prefix
	}

	return cfg
}
