package securitygroupattachment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-asc/terraform-provider-asc/internal/client"
	"github.com/stretchr/testify/assert"
)

func newAuthServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
}

func TestResource_Attach(t *testing.T) {
	authServer := newAuthServer()
	defer authServer.Close()

	tests := []struct {
		name           string
		sgName         string
		targetType     string
		targetName     string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:       "successful attach",
			sgName:     "my-sg",
			targetType: "VirtualMachine",
			targetName: "my-vm",
			mockResponse: &client.SecurityGroupAttachment{
				Name:  "my-vm-my-sg-attach",
				SGRef: "my-sg",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "attach API error - not found",
			sgName:         "nonexistent-sg",
			targetType:     "VirtualMachine",
			targetName:     "my-vm",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, fmt.Sprintf("/api/network/v1/namespaces/test-ns/security-groups/%s/attach", tt.sgName), r.URL.Path)

				var body client.AttachSecurityGroupRequest
				err := json.NewDecoder(r.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.targetType, body.TargetType)
				assert.Equal(t, tt.targetName, body.TargetName)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			networkClient := client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network

			sga, err := networkClient.AttachSecurityGroup(context.Background(), tt.sgName, tt.targetType, tt.targetName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "my-vm-my-sg-attach", sga.Name)
				assert.Equal(t, "my-sg", sga.SGRef)
			}
		})
	}
}

func TestResource_GetAttachment(t *testing.T) {
	authServer := newAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(&client.SecurityGroupAttachment{
				Name:  "my-vm-my-sg-attach",
				SGRef: "my-sg",
			})
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	networkClient := client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network

	sga, err := networkClient.GetSecurityGroupAttachment(context.Background(), "my-sg", "my-vm-my-sg-attach")
	assert.NoError(t, err)
	assert.Equal(t, "my-vm-my-sg-attach", sga.Name)
	assert.Equal(t, "my-sg", sga.SGRef)
}

func TestResource_Detach(t *testing.T) {
	authServer := newAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/network/v1/namespaces/test-ns/security-groups/my-sg/attachments/my-vm-my-sg-attach", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	networkClient := client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network

	err := networkClient.DetachSecurityGroup(context.Background(), "my-sg", "my-vm-my-sg-attach")
	assert.NoError(t, err)
}
