package client

import "os"

// ServiceEndpoint represents configuration for a service endpoint
type ServiceEndpoint struct {
	PathPrefix string
}

// ServiceConfig holds configuration for all DSPC API services
type ServiceConfig struct {
	VM            ServiceEndpoint
	VMGroup       ServiceEndpoint
	Network       ServiceEndpoint
	ManagedDB     ServiceEndpoint
	BlockStorage  ServiceEndpoint
	Authorization ServiceEndpoint
	Function      ServiceEndpoint
	Container     ServiceEndpoint
	FileStorage   ServiceEndpoint
	ObjectStorage ServiceEndpoint
	Cluster       ServiceEndpoint
}

// DefaultServiceConfig returns the default service configuration
// that matches the current DSPC API structure behind Envoy gateway
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		VM:            ServiceEndpoint{PathPrefix: "/api/vm"},
		VMGroup:       ServiceEndpoint{PathPrefix: "/api/vm"}, // Shares path with VM service
		Network:       ServiceEndpoint{PathPrefix: "/api/network"},
		ManagedDB:     ServiceEndpoint{PathPrefix: "/api/mdb"},
		BlockStorage:  ServiceEndpoint{PathPrefix: "/api/vm"}, // Shares path with VM service
		Authorization: ServiceEndpoint{PathPrefix: "/api/authorization"},
		Function:      ServiceEndpoint{PathPrefix: "/api/serverless-container"},
		Container:     ServiceEndpoint{PathPrefix: "/api/containers"},
		FileStorage:   ServiceEndpoint{PathPrefix: "/api/file"},
		ObjectStorage: ServiceEndpoint{PathPrefix: "/api/object-storage"},
		Cluster:       ServiceEndpoint{PathPrefix: "/api/cluster"},
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
	if prefix := os.Getenv("DSPC_MANAGED_DB_PATH_PREFIX"); prefix != "" {
		cfg.ManagedDB.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_STORAGE_PATH_PREFIX"); prefix != "" {
		cfg.BlockStorage.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_AUTH_PATH_PREFIX"); prefix != "" {
		cfg.Authorization.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_FUNCTION_PATH_PREFIX"); prefix != "" {
		cfg.Function.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_CONTAINER_PATH_PREFIX"); prefix != "" {
		cfg.Container.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_FILE_PATH_PREFIX"); prefix != "" {
		cfg.FileStorage.PathPrefix = prefix
	}
	if prefix := os.Getenv("DSPC_CLUSTER_PATH_PREFIX"); prefix != "" {
		cfg.Cluster.PathPrefix = prefix
	}

	return cfg
}
