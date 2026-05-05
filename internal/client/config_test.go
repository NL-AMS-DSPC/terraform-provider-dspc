package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	assert.Equal(t, "/api/vm", cfg.VM.PathPrefix)
	assert.Equal(t, "/api/network", cfg.Network.PathPrefix)
	assert.Equal(t, "/api/mdb", cfg.ManagedDB.PathPrefix)
	assert.Equal(t, "/api/vm", cfg.BlockStorage.PathPrefix, "BlockStorage should share path with VM")
}

func TestLoadServiceConfig_WithDefaults(t *testing.T) {
	// Ensure no env vars are set
	_ = os.Unsetenv("DSPC_VM_PATH_PREFIX")
	_ = os.Unsetenv("DSPC_NETWORK_PATH_PREFIX")
	_ = os.Unsetenv("DSPC_MDB_PATH_PREFIX")
	_ = os.Unsetenv("DSPC_STORAGE_PATH_PREFIX")

	cfg := LoadServiceConfig()

	assert.Equal(t, "/api/vm", cfg.VM.PathPrefix)
	assert.Equal(t, "/api/network", cfg.Network.PathPrefix)
	assert.Equal(t, "/api/mdb", cfg.ManagedDB.PathPrefix)
	assert.Equal(t, "/api/vm", cfg.BlockStorage.PathPrefix)
}

func TestLoadServiceConfig_WithEnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
		cleanup  func()
		validate func(*testing.T, ServiceConfig)
	}{
		{
			name: "override VM path prefix",
			setupEnv: func() {
				_ = os.Setenv("DSPC_VM_PATH_PREFIX", "/custom/vm")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_VM_PATH_PREFIX")
			},
			validate: func(t *testing.T, cfg ServiceConfig) {
				assert.Equal(t, "/custom/vm", cfg.VM.PathPrefix)
				assert.Equal(t, "/api/network", cfg.Network.PathPrefix)
				assert.Equal(t, "/api/mdb", cfg.ManagedDB.PathPrefix)
				assert.Equal(t, "/api/vm", cfg.BlockStorage.PathPrefix)
			},
		},
		{
			name: "override network path prefix",
			setupEnv: func() {
				_ = os.Setenv("DSPC_NETWORK_PATH_PREFIX", "/custom/network")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_NETWORK_PATH_PREFIX")
			},
			validate: func(t *testing.T, cfg ServiceConfig) {
				assert.Equal(t, "/api/vm", cfg.VM.PathPrefix)
				assert.Equal(t, "/custom/network", cfg.Network.PathPrefix)
				assert.Equal(t, "/api/mdb", cfg.ManagedDB.PathPrefix)
				assert.Equal(t, "/api/vm", cfg.BlockStorage.PathPrefix)
			},
		},
		{
			name: "override managed database path prefix",
			setupEnv: func() {
				_ = os.Setenv("DSPC_MDB_PATH_PREFIX", "/custom/mdb")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_MDB_PATH_PREFIX")
			},
			validate: func(t *testing.T, cfg ServiceConfig) {
				assert.Equal(t, "/api/vm", cfg.VM.PathPrefix)
				assert.Equal(t, "/api/network", cfg.Network.PathPrefix)
				assert.Equal(t, "/custom/mdb", cfg.ManagedDB.PathPrefix)
				assert.Equal(t, "/api/vm", cfg.BlockStorage.PathPrefix)
			},
		},
		{
			name: "override storage path prefix",
			setupEnv: func() {
				_ = os.Setenv("DSPC_STORAGE_PATH_PREFIX", "/custom/storage")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_STORAGE_PATH_PREFIX")
			},
			validate: func(t *testing.T, cfg ServiceConfig) {
				assert.Equal(t, "/api/vm", cfg.VM.PathPrefix)
				assert.Equal(t, "/api/network", cfg.Network.PathPrefix)
				assert.Equal(t, "/api/mdb", cfg.ManagedDB.PathPrefix)
				assert.Equal(t, "/custom/storage", cfg.BlockStorage.PathPrefix)
			},
		},
		{
			name: "override all path prefixes",
			setupEnv: func() {
				_ = os.Setenv("DSPC_VM_PATH_PREFIX", "/v2/vm")
				_ = os.Setenv("DSPC_NETWORK_PATH_PREFIX", "/v2/network")
				_ = os.Setenv("DSPC_MDB_PATH_PREFIX", "/v2/mdb")
				_ = os.Setenv("DSPC_STORAGE_PATH_PREFIX", "/v2/storage")
			},
			cleanup: func() {
				_ = os.Unsetenv("DSPC_VM_PATH_PREFIX")
				_ = os.Unsetenv("DSPC_NETWORK_PATH_PREFIX")
				_ = os.Unsetenv("DSPC_MDB_PATH_PREFIX")
				_ = os.Unsetenv("DSPC_STORAGE_PATH_PREFIX")
			},
			validate: func(t *testing.T, cfg ServiceConfig) {
				assert.Equal(t, "/v2/vm", cfg.VM.PathPrefix)
				assert.Equal(t, "/v2/network", cfg.Network.PathPrefix)
				assert.Equal(t, "/v2/mdb", cfg.ManagedDB.PathPrefix)
				assert.Equal(t, "/v2/storage", cfg.BlockStorage.PathPrefix)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanup()

			cfg := LoadServiceConfig()
			tt.validate(t, cfg)
		})
	}
}

func TestLoadServiceConfig_EmptyEnvVarsIgnored(t *testing.T) {
	// Set empty environment variables
	_ = os.Setenv("DSPC_VM_PATH_PREFIX", "")
	_ = os.Setenv("DSPC_NETWORK_PATH_PREFIX", "")
	_ = os.Setenv("DSPC_MDB_PATH_PREFIX", "")
	_ = os.Setenv("DSPC_STORAGE_PATH_PREFIX", "")
	defer func() {
		_ = os.Unsetenv("DSPC_VM_PATH_PREFIX")
		_ = os.Unsetenv("DSPC_NETWORK_PATH_PREFIX")
		_ = os.Unsetenv("DSPC_MDB_PATH_PREFIX")
		_ = os.Unsetenv("DSPC_STORAGE_PATH_PREFIX")
	}()

	cfg := LoadServiceConfig()

	// Empty env vars should be ignored, defaults should be used
	assert.Equal(t, "/api/vm", cfg.VM.PathPrefix)
	assert.Equal(t, "/api/network", cfg.Network.PathPrefix)
	assert.Equal(t, "/api/mdb", cfg.ManagedDB.PathPrefix)
	assert.Equal(t, "/api/vm", cfg.BlockStorage.PathPrefix)
}
