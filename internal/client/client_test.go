package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_ContextTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]*VM{{Name: "test-vm"}})
	}))
	defer server.Close()

	// Create client with short timeout
	client := NewDspcClient(server.URL, "test-ns", "test-api-key", 1) // 1 second timeout

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Test that context timeout is respected
	_, err := client.VirtualMachines.ListVMs(ctx)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check if error is context-related
	if !isContextError(err) {
		t.Errorf("Expected context error, got: %v", err)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]*VM{{Name: "test-vm"}})
	}))
	defer server.Close()

	// Create client
	client := NewDspcClient(server.URL, "test-ns", "test-api-key", 30)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Test that context cancellation is respected
	_, err := client.VirtualMachines.ListVMs(ctx)
	if err == nil {
		t.Error("Expected cancellation error, got nil")
	}

	// Check if error is context-related
	if !isContextError(err) {
		t.Errorf("Expected context error, got: %v", err)
	}
}

// Helper function to check if error is context-related
func isContextError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) ||
		strings.Contains(err.Error(), "context deadline exceeded") ||
		strings.Contains(err.Error(), "context canceled")
}
