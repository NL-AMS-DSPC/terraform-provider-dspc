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
				apiClient: newTestHttpClient(
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
				apiClient: newTestHttpClient(
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
				apiClient: newTestHttpClient(
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
				apiClient: newTestHttpClient(
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

func newTestHttpClient(responseTime int64, timeoutSeconds int64, resp interface{}) apiClient {
	client := newApiClient("https://example.com", "test-ns", "test-api-key", timeoutSeconds)
	client.httpClient.Transport = &recorderRoundTripper{
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response. Doesn't actually sleep 2 seconds.
			time.Sleep(time.Duration(responseTime) * time.Second)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}),
	}
	return client
}
