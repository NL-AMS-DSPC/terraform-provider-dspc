package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_RequestTimesOut(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		client := &DspcClient{
			VirtualMachines: &virtualMachineClient{
				apiClient: newTestHTTPClient(
					16,
					15,
					[]*VM{{Name: "test-vm"}}),
			},
		}

		ctx := context.Background()
		_, err := client.VirtualMachines.ListVMs(ctx)

		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// Wait the rest of the time to ensure the response goroutine finishes
		time.Sleep(15 * time.Second)
	})
}

func TestClient_SlowHttpResponse(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		client := &DspcClient{
			VirtualMachines: &virtualMachineClient{
				apiClient: newTestHTTPClient(
					14,
					15,
					[]*VM{{Name: "test-vm"}}),
			},
		}

		ctx := context.Background()
		vms, err := client.VirtualMachines.ListVMs(ctx)

		assert.NoError(t, err)
		assert.Len(t, vms, 1)

		// Wait the rest of the time to ensure the response goroutine finishes
		time.Sleep(15 * time.Second)
	})
}

func TestClient_ContextCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		client := &DspcClient{
			VirtualMachines: &virtualMachineClient{
				apiClient: newTestHTTPClient(
					10,
					15,
					[]*VM{{Name: "test-vm"}}),
			},
		}

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel context after short delay
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Test that context timeout is respected
		_, err := client.VirtualMachines.ListVMs(ctx)
		assert.ErrorIs(t, err, context.Canceled)

		// Wait the rest of the time to ensure the response goroutine finishes
		time.Sleep(15 * time.Second)
	})
}

func TestClient_ContextTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		client := &DspcClient{
			VirtualMachines: &virtualMachineClient{
				apiClient: newTestHTTPClient(
					10,
					15,
					[]*VM{{Name: "test-vm"}}),
			},
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test that context timeout is respected
		_, err := client.VirtualMachines.ListVMs(ctx)
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// Wait the rest of the time to ensure the response goroutine finishes
		time.Sleep(15 * time.Second)
	})
}

type recorderRoundTripper struct {
	handler http.Handler
}

func (r *recorderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resChan := make(chan *http.Response, 1)
	f := func() {
		rr := httptest.NewRecorder()
		r.handler.ServeHTTP(rr, req)
		resChan <- rr.Result()
	}

	go f()

	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case res := <-resChan:
		return res, nil
	}
}

// nolint:unparam // timeoutSeconds may be useful in future tests
func newTestHTTPClient(responseTime int64, timeoutSeconds int64, resp interface{}) apiClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &recorderRoundTripper{
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				// Simulate slow response. Doesn't actually sleep the given amount
				time.Sleep(time.Duration(responseTime) * time.Second)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			}),
		},
	}

	// Create a mock auth manager that returns a dummy token
	authMgr := &authManager{ // nolint:gosec
		httpClient:  httpClient,
		authURL:     "https://auth.example.com",
		org:         "test-realm",
		username:    "test-client-id",
		password:    "test-client-secret",
		accessToken: "mock-jwt-token",
		expiresAt:   time.Now().Add(1 * time.Hour),
	}

	cfg := DefaultServiceConfig()
	return newAPIClient("https://example.com", "test-ns", cfg.VM.PathPrefix, authMgr, httpClient)
}
