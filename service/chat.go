package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tomohiro-owada/gmn-gui/internal/api"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ChatMessage represents a message displayed in the UI
type ChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" | "model" | "tool_call" | "tool_result"
	Content   string    `json:"content"`
	ToolName  string    `json:"toolName,omitempty"`
	ToolArgs  string    `json:"toolArgs,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatStreamEvent is emitted to the frontend during streaming
type ChatStreamEvent struct {
	Type     string `json:"type"` // "start" | "content" | "tool_call" | "tool_result" | "done" | "error"
	Text     string `json:"text,omitempty"`
	ToolName string `json:"toolName,omitempty"`
	ToolArgs string `json:"toolArgs,omitempty"`
}

// ChatService handles multi-turn conversation with streaming
type ChatService struct {
	ctx      context.Context
	settings *SettingsService
	mcp      *MCPManager
	mu       sync.Mutex

	// Conversation state
	messages []ChatMessage  // UI display messages
	history  []api.Content  // API request history
	model    string         // Per-session model (overrides default)
	cancel   context.CancelFunc
}

// NewChatService creates a new chat service
func NewChatService(settings *SettingsService, mcp *MCPManager) *ChatService {
	return &ChatService{
		settings: settings,
		mcp:      mcp,
	}
}

// GetModel returns the current session model
func (c *ChatService) GetModel() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.model != "" {
		return c.model
	}
	return c.settings.GetDefaultModel()
}

// SetModel sets the model for the current session
func (c *ChatService) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model = model
}

// SetContext sets the Wails runtime context
func (c *ChatService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SendMessage sends a user message and starts streaming the response
func (c *ChatService) SendMessage(text string) error {
	c.mu.Lock()

	// Add user message to history
	userMsg := ChatMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   text,
		Timestamp: time.Now(),
	}
	c.messages = append(c.messages, userMsg)
	c.history = append(c.history, api.Content{
		Role:  "user",
		Parts: []api.Part{{Text: text}},
	})

	c.mu.Unlock()

	// Emit message update
	runtime.EventsEmit(c.ctx, "chat:messages", c.GetMessages())

	// Start streaming in goroutine
	go c.streamResponse()

	return nil
}

// StopGeneration cancels the current streaming generation
func (c *ChatService) StopGeneration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}

// ClearHistory clears all conversation history and resets to default model
func (c *ChatService) ClearHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = nil
	c.history = nil
	c.model = ""
	runtime.EventsEmit(c.ctx, "chat:messages", []ChatMessage{})
}

// GetMessages returns all messages for UI display
func (c *ChatService) GetMessages() []ChatMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]ChatMessage, len(c.messages))
	copy(result, c.messages)
	return result
}

func (c *ChatService) streamResponse() {
	ctx, cancel := context.WithCancel(c.ctx)
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.cancel = nil
		c.mu.Unlock()
		cancel()
	}()

	// Get authenticated API client
	client, err := c.settings.EnsureAuth(ctx)
	if err != nil {
		runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
			Type: "error",
			Text: fmt.Sprintf("Authentication failed: %v", err),
		})
		return
	}

	c.doStream(ctx, client)
}

func (c *ChatService) doStream(ctx context.Context, client *api.Client) {
	// Build tools from MCP
	var tools []api.Tool
	mcpTools := c.mcp.GetAllTools()
	if len(mcpTools) > 0 {
		tools = []api.Tool{{FunctionDeclarations: mcpTools}}
	}

	// Build request
	c.mu.Lock()
	historyCopy := make([]api.Content, len(c.history))
	copy(historyCopy, c.history)
	c.mu.Unlock()

	req := &api.GenerateRequest{
		Model:   c.GetModel(),
		Project: c.settings.GetProjectID(),
		Request: api.InnerRequest{
			Contents: historyCopy,
			Tools:    tools,
		},
	}

	// Start streaming
	events, err := client.GenerateStream(ctx, req)
	if err != nil {
		runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
			Type: "error",
			Text: fmt.Sprintf("Stream failed: %v", err),
		})
		return
	}

	runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{Type: "start"})

	var fullText string
	var pendingToolCalls []*api.FunctionCall

	for event := range events {
		switch event.Type {
		case "content":
			fullText += event.Text
			runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
				Type: "content",
				Text: event.Text,
			})

		case "tool_call":
			pendingToolCalls = append(pendingToolCalls, event.ToolCall)
			argsJSON, _ := json.Marshal(event.ToolCall.Args)
			runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
				Type:     "tool_call",
				ToolName: event.ToolCall.Name,
				ToolArgs: string(argsJSON),
			})

		case "error":
			runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
				Type: "error",
				Text: event.Error,
			})
			return

		case "done":
			// handled below
		}
	}

	// Add model response to history
	c.mu.Lock()
	if fullText != "" {
		c.messages = append(c.messages, ChatMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "model",
			Content:   fullText,
			Timestamp: time.Now(),
		})
	}

	// Build model parts for API history
	var modelParts []api.Part
	if fullText != "" {
		modelParts = append(modelParts, api.Part{Text: fullText})
	}
	for _, tc := range pendingToolCalls {
		modelParts = append(modelParts, api.Part{FunctionCall: tc})
	}
	if len(modelParts) > 0 {
		c.history = append(c.history, api.Content{
			Role:  "model",
			Parts: modelParts,
		})
	}
	c.mu.Unlock()

	// Handle tool calls if any
	if len(pendingToolCalls) > 0 {
		c.handleToolCalls(ctx, client, pendingToolCalls)
		return
	}

	runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{Type: "done"})
	runtime.EventsEmit(c.ctx, "chat:messages", c.GetMessages())
}

func (c *ChatService) handleToolCalls(ctx context.Context, client *api.Client, toolCalls []*api.FunctionCall) {
	var toolParts []api.Part

	for _, tc := range toolCalls {
		// Add tool call message to UI
		argsJSON, _ := json.Marshal(tc.Args)
		c.mu.Lock()
		c.messages = append(c.messages, ChatMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "tool_call",
			Content:   tc.Name,
			ToolName:  tc.Name,
			ToolArgs:  string(argsJSON),
			Timestamp: time.Now(),
		})
		c.mu.Unlock()

		// Execute tool via MCPManager
		result, err := c.mcp.CallTool(ctx, tc.Name, tc.Args)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}

		runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
			Type:     "tool_result",
			ToolName: tc.Name,
			Text:     result,
		})

		// Add tool result to UI messages
		c.mu.Lock()
		c.messages = append(c.messages, ChatMessage{
			ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			Role:      "tool_result",
			Content:   result,
			ToolName:  tc.Name,
			Timestamp: time.Now(),
		})
		c.mu.Unlock()

		// Build function response for API
		toolParts = append(toolParts, api.Part{
			FunctionResp: &api.FunctionResp{
				Name:     tc.Name,
				Response: map[string]interface{}{"result": result},
			},
		})
	}

	// Add tool results to API history
	c.mu.Lock()
	c.history = append(c.history, api.Content{
		Role:  "user",
		Parts: toolParts,
	})
	c.mu.Unlock()

	runtime.EventsEmit(c.ctx, "chat:messages", c.GetMessages())

	// Continue the conversation with tool results
	c.doStream(ctx, client)
}
