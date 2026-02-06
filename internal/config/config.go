// Package config provides configuration loading for geminimini.
// This file was modified from the original Gemini CLI.
// Copyright 2025 Google LLC
// Copyright 2025 Tomohiro Owada
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	geminiDir    = ".gemini"
	settingsFile = "settings.json"
)

// Config is the main configuration structure
type Config struct {
	Security   SecurityConfig             `json:"security"`
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
	General    GeneralConfig              `json:"general"`
	Output     OutputConfig               `json:"output"`
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	Auth AuthConfig `json:"auth"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	SelectedType string `json:"selectedType"`
}

// MCPServerConfig holds MCP server configuration
type MCPServerConfig struct {
	// Stdio transport
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	CWD     string            `json:"cwd,omitempty"`

	// HTTP/SSE transport
	URL     string            `json:"url,omitempty"`
	Type    string            `json:"type,omitempty"` // "sse" | "http"
	Headers map[string]string `json:"headers,omitempty"`

	// Common
	Timeout      int      `json:"timeout,omitempty"`
	Trust        bool     `json:"trust,omitempty"`
	IncludeTools []string `json:"includeTools,omitempty"`
	ExcludeTools []string `json:"excludeTools,omitempty"`
}

// GeneralConfig holds general settings
type GeneralConfig struct {
	PreviewFeatures bool `json:"previewFeatures"`
}

// OutputConfig holds output settings
type OutputConfig struct {
	Format string `json:"format"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Security: SecurityConfig{
			Auth: AuthConfig{
				SelectedType: "oauth-personal",
			},
		},
		MCPServers: make(map[string]MCPServerConfig),
		General: GeneralConfig{
			PreviewFeatures: false,
		},
		Output: OutputConfig{
			Format: "text",
		},
	}
}

// GeminiDir returns the path to ~/.gemini
func GeminiDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, geminiDir), nil
}

// Load loads the configuration from ~/.gemini/settings.json
func Load() (*Config, error) {
	geminiPath, err := GeminiDir()
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	// Load global settings
	globalPath := filepath.Join(geminiPath, settingsFile)
	if err := loadFile(globalPath, cfg); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Load project settings (optional, overrides global)
	cwd, err := os.Getwd()
	if err == nil {
		projectPath := filepath.Join(cwd, geminiDir, settingsFile)
		if err := loadFile(projectPath, cfg); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Load extensions (MCP servers from ~/.gemini/extensions/*)
	if err := loadExtensions(geminiPath, cwd, cfg); err != nil {
		// Non-fatal: just skip extensions
		_ = err
	}

	return cfg, nil
}

// geminiExtension represents a gemini-extension.json file
type geminiExtension struct {
	Name       string                    `json:"name"`
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// extensionEnablement maps extension name to enablement config
type extensionEnablement struct {
	Overrides []string `json:"overrides"`
}

func loadExtensions(geminiPath, cwd string, cfg *Config) error {
	extDir := filepath.Join(geminiPath, "extensions")

	// Load enablement config
	enablement := make(map[string]extensionEnablement)
	enablementPath := filepath.Join(extDir, "extension-enablement.json")
	if data, err := os.ReadFile(enablementPath); err == nil {
		_ = json.Unmarshal(data, &enablement)
	}

	// Scan extension directories
	entries, err := os.ReadDir(extDir)
	if err != nil {
		return nil // no extensions directory
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		extName := entry.Name()
		extPath := filepath.Join(extDir, extName)
		manifestPath := filepath.Join(extPath, "gemini-extension.json")

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var ext geminiExtension
		if err := json.Unmarshal(data, &ext); err != nil {
			continue
		}

		// Check enablement: if there's an entry for this extension, check if cwd matches
		if en, ok := enablement[extName]; ok {
			if !isEnabledForDir(cwd, en.Overrides) {
				continue
			}
		}

		// Merge MCP servers from extension
		for serverName, serverCfg := range ext.MCPServers {
			// Resolve ${extensionPath} in CWD
			if serverCfg.CWD == "${extensionPath}" {
				serverCfg.CWD = extPath
			}
			// Don't override user-configured servers
			if _, exists := cfg.MCPServers[serverName]; !exists {
				cfg.MCPServers[serverName] = serverCfg
			}
		}
	}

	return nil
}

func isEnabledForDir(cwd string, overrides []string) bool {
	if len(overrides) == 0 {
		return true // no overrides = enabled everywhere
	}
	for _, pattern := range overrides {
		matched, err := filepath.Match(pattern, cwd)
		if err == nil && matched {
			return true
		}
		// Also check if cwd is under the pattern directory (glob * at end)
		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			prefix := pattern[:len(pattern)-1]
			if len(cwd) >= len(prefix) && cwd[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

// CachedState represents cached state for geminimini
type CachedState struct {
	ProjectID string `json:"projectId,omitempty"`
	UserTier  string `json:"userTier,omitempty"`
}

// LoadCachedState loads the cached state from gmn_state.json
func LoadCachedState() (*CachedState, error) {
	geminiPath, err := GeminiDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(geminiPath, "gmn_state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CachedState{}, nil
		}
		return nil, err
	}

	var state CachedState
	if err := json.Unmarshal(data, &state); err != nil {
		return &CachedState{}, nil
	}

	return &state, nil
}

// SaveCachedState saves the cached state to gmn_state.json
func SaveCachedState(state *CachedState) error {
	geminiPath, err := GeminiDir()
	if err != nil {
		return err
	}

	path := filepath.Join(geminiPath, "gmn_state.json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
