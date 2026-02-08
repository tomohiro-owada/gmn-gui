package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// AskUserQuestion represents a question sent to the user via ask_user tool
type AskUserQuestion struct {
	Question string            `json:"question"`
	Header   string            `json:"header"`
	Type     string            `json:"type"` // "choice" | "text" | "yesno"
	Options  []AskUserOption   `json:"options,omitempty"`
}

// AskUserOption represents a choice option
type AskUserOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
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
	workDir  string         // Working directory for this session
	cancel   context.CancelFunc

	// ask_user channel
	askUserCh chan string

	// Plan mode
	planMode bool
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

// GetWorkDir returns the current working directory
func (c *ChatService) GetWorkDir() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.workDir
}

// SetWorkDir sets the working directory for the current session
func (c *ChatService) SetWorkDir(dir string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workDir = dir
}

// GetPlanMode returns whether plan mode is active
func (c *ChatService) GetPlanMode() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.planMode
}

// SetPlanMode toggles plan mode (read-only tools only)
func (c *ChatService) SetPlanMode(enabled bool) {
	c.mu.Lock()
	prev := c.planMode
	c.planMode = enabled
	hasHistory := len(c.history) > 0

	// Inject a notice into conversation history so the model knows the mode changed
	if hasHistory {
		if enabled && !prev {
			c.history = append(c.history, api.Content{
				Role:  "user",
				Parts: []api.Part{{Text: "[SYSTEM: Plan Mode has been ACTIVATED. From now on, only use read-only tools. Do NOT modify any files. Explain your plan instead of executing changes.]"}},
			})
		} else if !enabled && prev {
			c.history = append(c.history, api.Content{
				Role:  "user",
				Parts: []api.Part{{Text: "[SYSTEM: Plan Mode has been DEACTIVATED. All tools are now available. You may freely use write_file, replace, run_shell_command, and any other tools to make changes as requested.]"}},
			})
		}
	}
	c.mu.Unlock()
}

// SetContext sets the Wails runtime context
func (c *ChatService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SendMessage sends a user message and starts streaming the response
func (c *ChatService) SendMessage(text string) error {
	return c.SendMessageWithFiles(text, nil)
}

// AttachedFile represents a file attached to a message
type AttachedFile struct {
	Path     string `json:"path"`
	MimeType string `json:"mimeType"`
}

// SendMessageWithFiles sends a message with optional file attachments
func (c *ChatService) SendMessageWithFiles(text string, files []AttachedFile) error {
	c.mu.Lock()

	// Build parts: text + inline files
	parts := []api.Part{}
	if text != "" {
		parts = append(parts, api.Part{Text: text})
	}

	// Read and encode files
	for _, file := range files {
		data, err := readFileAsBase64(file.Path)
		if err != nil {
			c.mu.Unlock()
			return fmt.Errorf("failed to read file %s: %w", file.Path, err)
		}
		parts = append(parts, api.Part{
			InlineData: &api.InlineData{
				MimeType: file.MimeType,
				Data:     data,
			},
		})
	}

	// Add user message to history
	displayContent := text
	if len(files) > 0 {
		displayContent += fmt.Sprintf(" [%d file(s) attached]", len(files))
	}
	userMsg := ChatMessage{
		ID:        fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   displayContent,
		Timestamp: time.Now(),
	}
	c.messages = append(c.messages, userMsg)

	c.history = append(c.history, api.Content{
		Role:  "user",
		Parts: parts,
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
	c.workDir = ""
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

// AskUser sends questions to the frontend and blocks until the user responds
func (c *ChatService) AskUser(ctx context.Context, questions []AskUserQuestion) (string, error) {
	c.mu.Lock()
	c.askUserCh = make(chan string, 1)
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.askUserCh = nil
		c.mu.Unlock()
	}()

	// Emit event to frontend
	runtime.EventsEmit(c.ctx, "chat:ask_user", questions)

	// Wait for user response or cancellation
	select {
	case <-ctx.Done():
		return "User cancelled the question.", nil
	case answer := <-c.askUserCh:
		return answer, nil
	}
}

// SubmitAskUserResponse is called from the frontend when the user answers
func (c *ChatService) SubmitAskUserResponse(answer string) {
	c.mu.Lock()
	ch := c.askUserCh
	c.mu.Unlock()

	if ch != nil {
		ch <- answer
	}
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
	inPlanMode := c.GetPlanMode()

	// Build tools: built-in + MCP (filtered in plan mode)
	var allDecls []api.FunctionDecl
	if inPlanMode {
		allDecls = PlanModeToolDeclarations()
	} else {
		allDecls = BuiltinToolDeclarations()
		mcpTools := c.mcp.GetAllTools()
		allDecls = append(allDecls, mcpTools...)
	}
	var tools []api.Tool
	if len(allDecls) > 0 {
		tools = []api.Tool{{FunctionDeclarations: allDecls}}
	}

	// Build request
	c.mu.Lock()
	historyCopy := make([]api.Content, len(c.history))
	copy(historyCopy, c.history)
	c.mu.Unlock()

	// Build system instruction with environment context
	systemPrompt := BuildSystemPrompt(c.GetWorkDir())
	if inPlanMode {
		systemPrompt += "\n\n## PLAN MODE ACTIVE\nYou are in Plan Mode. Only use read-only tools to explore the codebase and design an implementation plan. Do NOT make any changes to files. Present your plan to the user for approval before proceeding."
	}
	systemInstruction := &api.Content{
		Parts: []api.Part{{Text: systemPrompt}},
	}

	req := &api.GenerateRequest{
		Model:   c.GetModel(),
		Project: c.settings.GetProjectID(),
		Request: api.InnerRequest{
			Contents:          historyCopy,
			SystemInstruction: systemInstruction,
			Tools:             tools,
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
	var thoughtText string
	var pendingToolParts []api.Part

	for event := range events {
		switch event.Type {
		case "thought":
			thoughtText += event.Text

		case "content":
			fullText += event.Text
			runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{
				Type: "content",
				Text: event.Text,
			})

		case "tool_call":
			pendingToolParts = append(pendingToolParts, api.Part{
				FunctionCall:     event.ToolCall,
				ThoughtSignature: event.ThoughtSignature,
			})
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

	// Build model parts for API history (preserve thought + thoughtSignature)
	var modelParts []api.Part
	if thoughtText != "" {
		modelParts = append(modelParts, api.Part{Thought: true, Text: thoughtText})
	}
	if fullText != "" {
		modelParts = append(modelParts, api.Part{Text: fullText})
	}
	modelParts = append(modelParts, pendingToolParts...)
	if len(modelParts) > 0 {
		c.history = append(c.history, api.Content{
			Role:  "model",
			Parts: modelParts,
		})
	}
	c.mu.Unlock()

	// Handle tool calls if any
	if len(pendingToolParts) > 0 {
		c.handleToolCalls(ctx, client, pendingToolParts)
		return
	}

	runtime.EventsEmit(c.ctx, "chat:stream", ChatStreamEvent{Type: "done"})
	runtime.EventsEmit(c.ctx, "chat:messages", c.GetMessages())
}

func (c *ChatService) handleToolCalls(ctx context.Context, client *api.Client, toolCallParts []api.Part) {
	var toolRespParts []api.Part

	for _, part := range toolCallParts {
		tc := part.FunctionCall
		if tc == nil {
			continue
		}

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

		// Plan mode guard: reject non-read-only tools
		var result string
		var err error
		if c.GetPlanMode() && !IsPlanModeTool(tc.Name) {
			result = fmt.Sprintf("Error: tool %q is not allowed in Plan Mode. Only read-only tools are available.", tc.Name)
		} else if tc.Name == "ask_user" {
			result, err = c.execAskUser(ctx, tc.Args)
		} else if IsBuiltinTool(tc.Name) {
			result, err = ExecuteBuiltinTool(ctx, c.GetWorkDir(), tc.Name, tc.Args, c.settings)
		} else {
			result, err = c.mcp.CallTool(ctx, tc.Name, tc.Args)
		}
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
		toolRespParts = append(toolRespParts, api.Part{
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
		Parts: toolRespParts,
	})
	c.mu.Unlock()

	runtime.EventsEmit(c.ctx, "chat:messages", c.GetMessages())

	// Continue the conversation with tool results
	c.doStream(ctx, client)
}

func (c *ChatService) execAskUser(ctx context.Context, args map[string]interface{}) (string, error) {
	questionsRaw, ok := args["questions"].([]interface{})
	if !ok || len(questionsRaw) == 0 {
		// Fallback: single question string
		if q, ok := args["question"].(string); ok && q != "" {
			questionsRaw = []interface{}{map[string]interface{}{"question": q, "header": "Question", "type": "text"}}
		} else {
			return "", fmt.Errorf("questions array is required")
		}
	}

	var questions []AskUserQuestion
	for _, qRaw := range questionsRaw {
		qMap, ok := qRaw.(map[string]interface{})
		if !ok {
			continue
		}
		q := AskUserQuestion{
			Question: stringVal(qMap, "question"),
			Header:   stringVal(qMap, "header"),
			Type:     stringVal(qMap, "type"),
		}
		if q.Type == "" {
			q.Type = "text"
		}
		if opts, ok := qMap["options"].([]interface{}); ok {
			for _, oRaw := range opts {
				if oMap, ok := oRaw.(map[string]interface{}); ok {
					q.Options = append(q.Options, AskUserOption{
						Label:       stringVal(oMap, "label"),
						Description: stringVal(oMap, "description"),
					})
				}
			}
		}
		questions = append(questions, q)
	}

	return c.AskUser(ctx, questions)
}

func stringVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

// readFileAsBase64 reads a file and returns its base64-encoded content
func readFileAsBase64(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// FileData represents a file sent from frontend
type FileData struct {
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64-encoded
}

// SaveFilesToTemp saves files to temporary directory and returns AttachedFile structs
func (c *ChatService) SaveFilesToTemp(files []FileData) ([]AttachedFile, error) {
	// Get temp directory
	tempDir := filepath.Join(os.TempDir(), "gmn-gui")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	result := make([]AttachedFile, len(files))
	for i, file := range files {
		// Decode base64
		data, err := base64.StdEncoding.DecodeString(file.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode file %s: %w", file.Filename, err)
		}

		// Create temp file path
		tempPath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename))

		// Write file
		if err := os.WriteFile(tempPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Filename, err)
		}

		result[i] = AttachedFile{
			Path:     tempPath,
			MimeType: file.MimeType,
		}
	}

	return result, nil
}
