// Package api provides a client for the Gemini API.
// Copyright 2025 Tomohiro Owada
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// Code Assist API endpoint (same as official Gemini CLI)
	baseURL    = "https://cloudcode-pa.googleapis.com"
	apiVersion = "v1internal"
)

// Client is a Gemini API client
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new API client
func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// GenerateRequest is a request to generate content (Code Assist API format)
type GenerateRequest struct {
	Model        string       `json:"model"`
	Project      string       `json:"project,omitempty"`
	UserPromptID string       `json:"user_prompt_id,omitempty"`
	Request      InnerRequest `json:"request"`
}

// InnerRequest is the inner request structure for Code Assist API
type InnerRequest struct {
	Contents []Content        `json:"contents"`
	Config   GenerationConfig `json:"generationConfig,omitempty"`
	Tools    []Tool           `json:"tools,omitempty"`
}

// Content represents a message content
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part represents a content part
type Part struct {
	Text         string        `json:"text,omitempty"`
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
	FunctionResp *FunctionResp `json:"functionResponse,omitempty"`
}

// FunctionCall represents a tool call
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// FunctionResp represents a tool response
type FunctionResp struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// GenerationConfig holds generation parameters
type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

// Tool represents a tool definition
type Tool struct {
	FunctionDeclarations []FunctionDecl `json:"functionDeclarations"`
}

// FunctionDecl represents a function declaration
type FunctionDecl struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// GenerateResponse is a response from generate content (Code Assist API format)
type GenerateResponse struct {
	Response InnerResponse `json:"response"`
	TraceID  string        `json:"traceId,omitempty"`
}

// InnerResponse is the inner response structure for Code Assist API
type InnerResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

// UsageMetadata holds token usage information
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// Generate sends a non-streaming generate request
func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	endpoint := fmt.Sprintf("%s/%s:generateContent", c.baseURL, apiVersion)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type       string         `json:"type"`
	Model      string         `json:"model,omitempty"`
	Text       string         `json:"text,omitempty"`
	ToolCall   *FunctionCall  `json:"tool_call,omitempty"`
	ToolResult *ToolResult    `json:"tool_result,omitempty"`
	Usage      *UsageMetadata `json:"usage,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// ToolResult represents a tool execution result
type ToolResult struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
}

// LoadCodeAssistRequest is the request to load user's Code Assist status
type LoadCodeAssistRequest struct {
	CloudAICompanionProject string         `json:"cloudaicompanionProject,omitempty"`
	Metadata                ClientMetadata `json:"metadata"`
}

// ClientMetadata represents client metadata for Code Assist API
type ClientMetadata struct {
	IdeType    string `json:"ideType,omitempty"`
	Platform   string `json:"platform,omitempty"`
	PluginType string `json:"pluginType,omitempty"`
}

// LoadCodeAssistResponse is the response from loadCodeAssist
type LoadCodeAssistResponse struct {
	CurrentTier             *UserTier        `json:"currentTier,omitempty"`
	AllowedTiers            []UserTier       `json:"allowedTiers,omitempty"`
	IneligibleTiers         []IneligibleTier `json:"ineligibleTiers,omitempty"`
	CloudAICompanionProject string           `json:"cloudaicompanionProject,omitempty"`
}

// UserTier represents a user's tier
type UserTier struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// IneligibleTier represents a tier the user is not eligible for
type IneligibleTier struct {
	ReasonCode    string `json:"reasonCode"`
	ReasonMessage string `json:"reasonMessage"`
	TierID        string `json:"tierId"`
	TierName      string `json:"tierName"`
	ValidationURL string `json:"validationUrl,omitempty"`
}

// LoadCodeAssist loads the user's Code Assist status and returns the project ID
func (c *Client) LoadCodeAssist(ctx context.Context) (*LoadCodeAssistResponse, error) {
	endpoint := fmt.Sprintf("%s/%s:loadCodeAssist", c.baseURL, apiVersion)

	req := LoadCodeAssistRequest{
		Metadata: ClientMetadata{
			IdeType:    "GEMINI_CLI",
			Platform:   "PLATFORM_UNSPECIFIED",
			PluginType: "GEMINI",
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result LoadCodeAssistResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GenerateStream sends a streaming generate request
func (c *Client) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan StreamEvent, error) {
	endpoint := fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse", c.baseURL, apiVersion)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	events := make(chan StreamEvent)

	go func() {
		defer close(events)
		defer resp.Body.Close()

		// Send start event
		events <- StreamEvent{Type: "start", Model: req.Model}

		reader := bufio.NewReader(resp.Body)
		var usage *UsageMetadata

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					events <- StreamEvent{Type: "error", Error: err.Error()}
				}
				break
			}

			line = strings.TrimSpace(line)
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk GenerateResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			// Store usage for final event
			if chunk.Response.UsageMetadata.TotalTokenCount > 0 {
				usage = &chunk.Response.UsageMetadata
			}

			// Extract text from candidates
			for _, candidate := range chunk.Response.Candidates {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						events <- StreamEvent{Type: "content", Text: part.Text}
					}
					if part.FunctionCall != nil {
						events <- StreamEvent{Type: "tool_call", ToolCall: part.FunctionCall}
					}
				}
			}
		}

		// Send done event
		events <- StreamEvent{Type: "done", Usage: usage}
	}()

	return events, nil
}
