package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/tomohiro-owada/gmn-gui/internal/api"
	"github.com/tomohiro-owada/gmn-gui/internal/config"
	"github.com/tomohiro-owada/gmn-gui/internal/mcp"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// MCPServerStatus represents the status of an MCP server for the UI
type MCPServerStatus struct {
	Name      string   `json:"name"`
	Connected bool     `json:"connected"`
	Command   string   `json:"command,omitempty"`
	URL       string   `json:"url,omitempty"`
	ToolCount int      `json:"toolCount"`
	Tools     []string `json:"tools,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// MCPManager manages multiple MCP server connections
type MCPManager struct {
	ctx      context.Context
	settings *SettingsService
	mu       sync.RWMutex
	clients  map[string]*mcp.Client
	errors   map[string]string
}

// NewMCPManager creates a new MCP manager
func NewMCPManager(settings *SettingsService) *MCPManager {
	return &MCPManager{
		settings: settings,
		clients:  make(map[string]*mcp.Client),
		errors:   make(map[string]string),
	}
}

// SetContext sets the Wails runtime context
func (m *MCPManager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// ListServers returns all configured MCP servers with their status
func (m *MCPManager) ListServers() []MCPServerStatus {
	cfg := m.settings.GetConfig()
	if cfg == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var servers []MCPServerStatus
	for name, serverCfg := range cfg.MCPServers {
		status := MCPServerStatus{
			Name:    name,
			Command: serverCfg.Command,
			URL:     serverCfg.URL,
		}

		if client, ok := m.clients[name]; ok {
			status.Connected = true
			status.ToolCount = len(client.Tools)
			for _, tool := range client.Tools {
				status.Tools = append(status.Tools, tool.Name)
			}
		}

		if errMsg, ok := m.errors[name]; ok {
			status.Error = errMsg
		}

		servers = append(servers, status)
	}

	return servers
}

// ConnectServer connects to a specific MCP server
func (m *MCPManager) ConnectServer(name string) error {
	cfg := m.settings.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	serverCfg, ok := cfg.MCPServers[name]
	if !ok {
		return fmt.Errorf("server %q not found in config", name)
	}

	m.mu.Lock()
	// Disconnect existing connection if any
	if existing, ok := m.clients[name]; ok {
		existing.Close()
		delete(m.clients, name)
	}
	delete(m.errors, name)
	m.mu.Unlock()

	if serverCfg.Command == "" {
		return fmt.Errorf("server %q has no command configured (HTTP transport not yet supported)", name)
	}

	client, err := mcp.NewClient(serverCfg.Command, serverCfg.Args, serverCfg.Env, serverCfg.CWD)
	if err != nil {
		m.mu.Lock()
		m.errors[name] = err.Error()
		m.mu.Unlock()
		return fmt.Errorf("failed to start server %q: %w", name, err)
	}

	ctx, cancel := context.WithTimeout(m.ctx, 30_000_000_000) // 30 seconds
	defer cancel()

	if err := client.Initialize(ctx); err != nil {
		client.Close()
		m.mu.Lock()
		m.errors[name] = err.Error()
		m.mu.Unlock()
		return fmt.Errorf("failed to initialize server %q: %w", name, err)
	}

	m.mu.Lock()
	m.clients[name] = client
	m.mu.Unlock()

	runtime.EventsEmit(m.ctx, "mcp:updated", m.ListServers())
	return nil
}

// DisconnectServer disconnects from a specific MCP server
func (m *MCPManager) DisconnectServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[name]
	if !ok {
		return nil
	}

	client.Close()
	delete(m.clients, name)
	delete(m.errors, name)

	runtime.EventsEmit(m.ctx, "mcp:updated", m.ListServers())
	return nil
}

// ConnectAll connects to all configured MCP servers
func (m *MCPManager) ConnectAll() {
	cfg := m.settings.GetConfig()
	if cfg == nil {
		return
	}

	for name := range cfg.MCPServers {
		if err := m.ConnectServer(name); err != nil {
			fmt.Printf("MCP: failed to connect %q: %v\n", name, err)
		}
	}
}

// DisconnectAll disconnects from all MCP servers
func (m *MCPManager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, client := range m.clients {
		client.Close()
		delete(m.clients, name)
	}
}

// GetAllTools returns tools from all connected servers as API function declarations.
// Tool names are prefixed with the server name to avoid conflicts (e.g. "myserver__toolname").
func (m *MCPManager) GetAllTools() []api.FunctionDecl {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tools []api.FunctionDecl
	for serverName, client := range m.clients {
		for _, tool := range client.Tools {
			params := sanitizeSchemaRaw(tool.InputSchema)
			// Sanitize name: Gemini API only allows [a-zA-Z0-9_]
			safeName := sanitizeToolName(serverName) + "__" + sanitizeToolName(tool.Name)
			tools = append(tools, api.FunctionDecl{
				Name:        safeName,
				Description: tool.Description,
				Parameters:  params,
			})
		}
	}
	return tools
}

// sanitizeToolName replaces characters not allowed in Gemini API function names with underscores.
func sanitizeToolName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// sanitizeSchemaRaw removes fields like $schema from a JSON schema (json.RawMessage).
func sanitizeSchemaRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw
	}
	sanitizeMap(m)
	out, err := json.Marshal(m)
	if err != nil {
		return raw
	}
	return out
}

func sanitizeMap(m map[string]interface{}) {
	delete(m, "$schema")
	for _, v := range m {
		if nested, ok := v.(map[string]interface{}); ok {
			sanitizeMap(nested)
		}
	}
}

// CallTool calls a tool on the appropriate MCP server.
// The toolName should be prefixed with the server name (e.g. "myserver__toolname").
func (m *MCPManager) CallTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	// Build lookup: sanitized server name → original server name + client
	m.mu.RLock()
	var matchedClient *mcp.Client
	var actualTool string

	for serverName, client := range m.clients {
		safeServer := sanitizeToolName(serverName)
		prefix := safeServer + "__"
		if strings.HasPrefix(toolName, prefix) {
			// Find the actual tool name by matching sanitized tool names
			remainder := toolName[len(prefix):]
			for _, tool := range client.Tools {
				if sanitizeToolName(tool.Name) == remainder {
					matchedClient = client
					actualTool = tool.Name
					break
				}
			}
			if matchedClient != nil {
				break
			}
		}
	}
	m.mu.RUnlock()

	if matchedClient == nil {
		return "", fmt.Errorf("MCP tool %q not found", toolName)
	}

	return matchedClient.CallTool(ctx, actualTool, args)
}

// AddServer adds a new MCP server to the configuration
func (m *MCPManager) AddServer(name string, command string, args string, env string) error {
	cfg := m.settings.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	var argsList []string
	if args != "" {
		if err := json.Unmarshal([]byte(args), &argsList); err != nil {
			// Treat as single arg if not JSON array
			argsList = strings.Fields(args)
		}
	}

	var envMap map[string]string
	if env != "" {
		envMap = make(map[string]string)
		if err := json.Unmarshal([]byte(env), &envMap); err != nil {
			return fmt.Errorf("invalid env format (expected JSON object): %w", err)
		}
	}

	cfg.MCPServers[name] = config.MCPServerConfig{
		Command: command,
		Args:    argsList,
		Env:     envMap,
	}

	// TODO: persist to settings.json
	runtime.EventsEmit(m.ctx, "mcp:updated", m.ListServers())
	return nil
}

// RemoveServer removes an MCP server from the configuration
func (m *MCPManager) RemoveServer(name string) error {
	// Disconnect first
	_ = m.DisconnectServer(name)

	cfg := m.settings.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	delete(cfg.MCPServers, name)

	// TODO: persist to settings.json
	runtime.EventsEmit(m.ctx, "mcp:updated", m.ListServers())
	return nil
}
