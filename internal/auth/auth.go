// Package auth provides OAuth authentication for geminimini.
// This file was modified from the original Gemini CLI.
// Copyright 2025 Google LLC
// Copyright 2025 Tomohiro Owada
// SPDX-License-Identifier: Apache-2.0
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tomohiro-owada/gmn-gui/internal/config"
)

const (
	oauthFile     = "oauth_creds.json"
	tokenEndpoint = "https://oauth2.googleapis.com/token"

	// OAuth credentials from official Gemini CLI
	// It's OK to embed these in source code for installed applications.
	// See: https://developers.google.com/identity/protocols/oauth2#installed
	clientID     = "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com"
	clientSecret = "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"
)

// Credentials holds OAuth token information
type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
	ExpiryDate   int64  `json:"expiry_date,omitempty"`
}

// IsExpired checks if the token is expired (with 5 minute margin)
func (c *Credentials) IsExpired() bool {
	if c.ExpiryDate == 0 {
		return false
	}
	// ExpiryDate is in milliseconds
	expiryTime := time.UnixMilli(c.ExpiryDate)
	return time.Now().Add(5 * time.Minute).After(expiryTime)
}

// Manager handles OAuth authentication
type Manager struct {
	geminiDir string
}

// NewManager creates a new auth manager
func NewManager() (*Manager, error) {
	geminiDir, err := config.GeminiDir()
	if err != nil {
		return nil, err
	}
	return &Manager{geminiDir: geminiDir}, nil
}

// LoadCredentials loads OAuth credentials from file or keychain
func (m *Manager) LoadCredentials() (*Credentials, error) {
	// Try keychain first (macOS only)
	creds, err := m.loadFromKeychain()
	if err == nil && creds != nil {
		return creds, nil
	}

	// Fall back to file
	return m.loadFromFile()
}

// loadFromFile reads credentials from oauth_creds.json
func (m *Manager) loadFromFile() (*Credentials, error) {
	path := filepath.Join(m.geminiDir, oauthFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials not found: run 'gemini' to authenticate first")
		}
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// RefreshToken refreshes an expired access token
func (m *Manager) RefreshToken(creds *Credentials) (*Credentials, error) {
	if creds.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available: run 'gemini' to re-authenticate")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", creds.RefreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.Post(
		tokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (status %d): run 'gemini' to re-authenticate", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &Credentials{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: creds.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiryDate:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).UnixMilli(),
	}, nil
}

// HTTPClient returns an HTTP client with the access token
func (m *Manager) HTTPClient(creds *Credentials) *http.Client {
	return &http.Client{
		Transport: &authTransport{
			token: creds.AccessToken,
			base:  http.DefaultTransport,
		},
	}
}

// authTransport adds Authorization header to requests
type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

const (
	authEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	oauthScope   = "openid email https://www.googleapis.com/auth/cloud-platform"
)

// StartLogin starts an OAuth login flow by opening the browser and waiting for the callback.
func (m *Manager) StartLogin(ctx context.Context) (*Credentials, error) {
	// Start a local server on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	// Generate state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Build consent URL
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", oauthScope)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("state", state)
	consentURL := authEndpoint + "?" + params.Encode()

	// Channel to receive the authorization code
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("OAuth error: %s", errMsg)
			fmt.Fprintf(w, "<html><body><h2>Authentication failed: %s</h2><p>You can close this window.</p></body></html>", errMsg)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code", http.StatusBadRequest)
			return
		}
		codeCh <- code
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Authentication successful!</h2><p>You can close this window and return to gmn-gui.</p></body></html>")
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Shutdown(context.Background())

	// Open the browser
	if err := openBrowser(consentURL); err != nil {
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for the callback or context cancellation
	var code string
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case code = <-codeCh:
	}

	// Exchange authorization code for tokens
	return m.exchangeCode(code, redirectURI)
}

func (m *Manager) exchangeCode(code, redirectURI string) (*Credentials, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	resp, err := http.Post(
		tokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (status %d)", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &Credentials{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Scope:        tokenResp.Scope,
		ExpiryDate:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).UnixMilli(),
	}, nil
}

// SaveCredentials saves credentials to the oauth file
func (m *Manager) SaveCredentials(creds *Credentials) error {
	path := filepath.Join(m.geminiDir, oauthFile)
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DeleteCredentials removes the stored credentials file
func (m *Manager) DeleteCredentials() error {
	path := filepath.Join(m.geminiDir, oauthFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
