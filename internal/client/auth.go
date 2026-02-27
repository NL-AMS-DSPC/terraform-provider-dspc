package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// authManager handles OAuth2 client credentials authentication and token caching
type authManager struct {
	httpClient  *http.Client
	authURL     string
	org         string
	username    string
	password    string
	accessToken string
	expiresAt   time.Time
	mu          sync.RWMutex
}

// tokenResponse represents the OAuth2 token response from Keycloak
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// newAuthManager creates a new authentication manager
func newAuthManager(httpClient *http.Client, authURL, org, username, password string) *authManager {
	return &authManager{
		httpClient: httpClient,
		authURL:    authURL,
		org:        org,
		username:   username,
		password:   password,
	}
}

// getToken returns a valid access token, refreshing it if necessary
func (a *authManager) getToken(ctx context.Context) (string, error) {
	a.mu.RLock()
	if a.accessToken != "" && time.Now().Before(a.expiresAt.Add(-minTokenLifetime)) {
		token := a.accessToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	// Token expired or doesn't exist, get a new one
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring write lock
	if a.accessToken != "" && time.Now().Before(a.expiresAt.Add(-minTokenLifetime)) {
		return a.accessToken, nil
	}

	// Request new token
	token, expiresIn, err := a.requestToken(ctx)
	if err != nil {
		return "", err
	}

	a.accessToken = token
	a.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return token, nil
}

// requestToken requests a new access token from the OAuth2 server
func (a *authManager) requestToken(ctx context.Context) (string, int64, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", a.authURL, a.org)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", a.username)
	data.Set("client_secret", a.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to request token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}
