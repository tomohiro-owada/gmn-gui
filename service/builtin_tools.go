package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomohiro-owada/gmn-gui/internal/api"
)

// builtinToolNames lists all built-in tool names for routing
var builtinToolNames = map[string]bool{
	"run_shell_command":  true,
	"read_file":          true,
	"read_many_files":    true,
	"write_file":         true,
	"replace":            true,
	"list_directory":     true,
	"glob":               true,
	"grep_search":        true,
	"google_web_search":  true,
	"web_fetch":          true,
	"write_todos":        true,
	"save_memory":        true,
	"ask_user":           true,
	"activate_skill":     true,
	"get_internal_docs":  true,
}

// IsBuiltinTool returns true if the tool name is a built-in tool
func IsBuiltinTool(name string) bool {
	return builtinToolNames[name]
}

// BuiltinToolDeclarations returns the function declarations for all built-in tools
func BuiltinToolDeclarations() []api.FunctionDecl {
	return []api.FunctionDecl{
		{
			Name:        "run_shell_command",
			Description: "Executes a shell command as `bash -c <command>`. Returns the combined stdout/stderr output, exit code if non-zero, and any error information. Use this for running build commands, tests, git operations, and other CLI tasks.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"command": {
						"type": "string",
						"description": "The bash command to execute."
					},
					"description": {
						"type": "string",
						"description": "Brief description of what the command does."
					},
					"dir_path": {
						"type": "string",
						"description": "Optional: Directory to run the command in. Defaults to the working directory."
					}
				},
				"required": ["command"]
			}`),
		},
		{
			Name:        "read_file",
			Description: "Reads and returns the content of a specified file. For text files, it can read specific line ranges using offset and limit parameters.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"file_path": {
						"type": "string",
						"description": "The absolute path to the file to read."
					},
					"offset": {
						"type": "number",
						"description": "Optional: The 0-based line number to start reading from."
					},
					"limit": {
						"type": "number",
						"description": "Optional: Maximum number of lines to read."
					}
				},
				"required": ["file_path"]
			}`),
		},
		{
			Name:        "write_file",
			Description: "Writes content to a specified file. Creates the file if it doesn't exist, or overwrites it if it does. Creates parent directories as needed.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"file_path": {
						"type": "string",
						"description": "The absolute path to the file to write."
					},
					"content": {
						"type": "string",
						"description": "The content to write to the file."
					}
				},
				"required": ["file_path", "content"]
			}`),
		},
		{
			Name:        "replace",
			Description: "Replaces text within a file. Finds the exact literal old_string and replaces it with new_string. Always read the file first to get the exact text to replace. Include enough context (at least 3 lines before and after) to uniquely identify the location.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"file_path": {
						"type": "string",
						"description": "The absolute path to the file to modify."
					},
					"old_string": {
						"type": "string",
						"description": "The exact literal text to replace. Must match exactly including whitespace and indentation."
					},
					"new_string": {
						"type": "string",
						"description": "The replacement text."
					},
					"expected_replacements": {
						"type": "number",
						"description": "Optional: Number of replacements expected. Defaults to 1."
					}
				},
				"required": ["file_path", "old_string", "new_string"]
			}`),
		},
		{
			Name:        "list_directory",
			Description: "Lists files and subdirectories in a specified directory. Returns names with a trailing / for directories.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"dir_path": {
						"type": "string",
						"description": "The absolute path to the directory to list."
					}
				},
				"required": ["dir_path"]
			}`),
		},
		{
			Name:        "glob",
			Description: "Finds files matching a glob pattern (e.g. '**/*.ts', 'src/**/*.go'). Returns absolute paths sorted by modification time (newest first).",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"pattern": {
						"type": "string",
						"description": "The glob pattern to match (e.g. '**/*.py', 'docs/*.md')."
					},
					"dir_path": {
						"type": "string",
						"description": "Optional: Directory to search within. Defaults to working directory."
					}
				},
				"required": ["pattern"]
			}`),
		},
		{
			Name:        "grep_search",
			Description: "Searches for a regex pattern within file contents. Returns matching lines with file paths and line numbers. Max 100 matches.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"pattern": {
						"type": "string",
						"description": "The regex pattern to search for."
					},
					"dir_path": {
						"type": "string",
						"description": "Optional: Directory to search within. Defaults to working directory."
					},
					"include": {
						"type": "string",
						"description": "Optional: Glob pattern to filter files (e.g. '*.js', '*.{ts,tsx}')."
					}
				},
				"required": ["pattern"]
			}`),
		},
		{
			Name:        "save_memory",
			Description: "Saves a fact or user preference to long-term memory (GEMINI.md). Use when the user asks you to remember something.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"fact": {
						"type": "string",
						"description": "The fact or preference to remember."
					}
				},
				"required": ["fact"]
			}`),
		},
		{
			Name:        "read_many_files",
			Description: "Reads content from multiple files specified by glob patterns within the working directory. Concatenates text file content into a single string. Use this when you need to read many files at once.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"include": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Glob patterns or paths to include (e.g. ['src/**/*.ts', 'README.md'])."
					},
					"exclude": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Optional: Glob patterns to exclude."
					}
				},
				"required": ["include"]
			}`),
		},
		{
			Name:        "google_web_search",
			Description: "Performs a web search using Google and returns results. Useful for finding information on the internet based on a query.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The search query to find information on the web."
					}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        "web_fetch",
			Description: "Fetches and returns the content of a URL. Supports HTTP and HTTPS URLs. Use this to retrieve web pages, API responses, or other online content.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"url": {
						"type": "string",
						"description": "The URL to fetch content from. Must start with http:// or https://."
					},
					"prompt": {
						"type": "string",
						"description": "Optional: Instructions on how to process or summarize the fetched content."
					}
				},
				"required": ["url"]
			}`),
		},
		{
			Name:        "write_todos",
			Description: "Manages a list of todo items for tracking subtasks. Use this to break complex tasks into manageable subtasks with status tracking.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"todos": {
						"type": "array",
						"description": "The complete list of todo items. This will replace the existing list.",
						"items": {
							"type": "object",
							"properties": {
								"description": {
									"type": "string",
									"description": "The description of the task."
								},
								"status": {
									"type": "string",
									"description": "The current status of the task.",
									"enum": ["pending", "in_progress", "completed", "cancelled"]
								}
							},
							"required": ["description", "status"]
						}
					}
				},
				"required": ["todos"]
			}`),
		},
		{
			Name:        "ask_user",
			Description: "Ask the user one or more questions to gather preferences, clarify requirements, or make decisions. Use when you need user input before proceeding.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"questions": {
						"type": "array",
						"description": "Questions to ask the user (1-4 questions).",
						"items": {
							"type": "object",
							"properties": {
								"question": {
									"type": "string",
									"description": "The question to ask. Should be clear and end with a question mark."
								},
								"header": {
									"type": "string",
									"description": "Very short label (max 16 chars). E.g. 'Auth method', 'Library'."
								},
								"type": {
									"type": "string",
									"description": "Question type: 'choice' for multiple-choice, 'text' for free-form, 'yesno' for confirmation.",
									"enum": ["choice", "text", "yesno"]
								},
								"options": {
									"type": "array",
									"description": "Choices for 'choice' type questions.",
									"items": {
										"type": "object",
										"properties": {
											"label": {"type": "string", "description": "Option display text (1-5 words)."},
											"description": {"type": "string", "description": "Brief explanation of this option."}
										},
										"required": ["label", "description"]
									}
								}
							},
							"required": ["question", "header"]
						}
					}
				},
				"required": ["questions"]
			}`),
		},
		{
			Name:        "activate_skill",
			Description: "Activates a specialized skill by name. Returns the skill's instructions. Skills are loaded from .gmn/skills/ in the project or ~/.gmn/skills/ globally. If no name is given, lists available skills.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The name of the skill to activate."
					}
				},
				"required": ["name"]
			}`),
		},
		{
			Name:        "get_internal_docs",
			Description: "Returns content of documentation files from the project's docs/ directory. If no path is provided, lists all available documentation paths.",
			Parameters: jsonRaw(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "Relative path to a doc file (e.g. 'cli/commands.md'). If omitted, lists all available docs."
					}
				}
			}`),
		},
	}
}

// planModeTools lists tool names allowed in plan mode (read-only)
var planModeTools = map[string]bool{
	"glob":               true,
	"grep_search":        true,
	"read_file":          true,
	"read_many_files":    true,
	"list_directory":     true,
	"google_web_search":  true,
	"web_fetch":          true,
	"ask_user":           true,
	"get_internal_docs":  true,
}

// IsPlanModeTool returns true if the tool is allowed in plan mode
func IsPlanModeTool(name string) bool {
	return planModeTools[name]
}

// PlanModeToolDeclarations returns only read-only tool declarations for plan mode
func PlanModeToolDeclarations() []api.FunctionDecl {
	all := BuiltinToolDeclarations()
	var filtered []api.FunctionDecl
	for _, decl := range all {
		if planModeTools[decl.Name] {
			filtered = append(filtered, decl)
		}
	}
	return filtered
}

func jsonRaw(s string) json.RawMessage {
	return json.RawMessage(s)
}

// ExecuteBuiltinTool runs a built-in tool and returns the result
func ExecuteBuiltinTool(ctx context.Context, workDir string, name string, args map[string]interface{}) (string, error) {
	switch name {
	case "run_shell_command":
		return execShellCommand(ctx, workDir, args)
	case "read_file":
		return execReadFile(args)
	case "read_many_files":
		return execReadManyFiles(workDir, args)
	case "write_file":
		return execWriteFile(args)
	case "replace":
		return execReplace(args)
	case "list_directory":
		return execListDirectory(args)
	case "glob":
		return execGlob(workDir, args)
	case "grep_search":
		return execGrepSearch(ctx, workDir, args)
	case "google_web_search":
		return execGoogleWebSearch(ctx, args)
	case "web_fetch":
		return execWebFetch(ctx, args)
	case "write_todos":
		return execWriteTodos(workDir, args)
	case "save_memory":
		return execSaveMemory(workDir, args)
	case "ask_user":
		return "", fmt.Errorf("ask_user should be handled by ChatService directly")
	case "activate_skill":
		return execActivateSkill(workDir, args)
	case "get_internal_docs":
		return execGetInternalDocs(workDir, args)
	default:
		return "", fmt.Errorf("unknown built-in tool: %s", name)
	}
}

// --- Tool implementations ---

func execShellCommand(ctx context.Context, workDir string, args map[string]interface{}) (string, error) {
	command, _ := args["command"].(string)
	if command == "" {
		return "", fmt.Errorf("command is required")
	}

	dir := workDir
	if d, ok := args["dir_path"].(string); ok && d != "" {
		dir = d
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "bash", "-c", command)
	if dir != "" {
		cmd.Dir = dir
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	result := out.String()
	if len(result) > 50000 {
		result = result[:50000] + "\n... (output truncated)"
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Sprintf("%s\nExit code: %d", result, exitErr.ExitCode()), nil
		}
		return result, fmt.Errorf("command error: %w", err)
	}

	if result == "" {
		result = "(empty output)"
	}

	return result, nil
}

func execReadFile(args map[string]interface{}) (string, error) {
	filePath, _ := args["file_path"].(string)
	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Handle offset/limit
	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	limit := len(lines)
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	if offset > 0 || limit < len(lines) {
		end := offset + limit
		if offset >= len(lines) {
			return "(offset beyond end of file)", nil
		}
		if end > len(lines) {
			end = len(lines)
		}
		lines = lines[offset:end]

		// Add line numbers
		var sb strings.Builder
		for i, line := range lines {
			sb.WriteString(fmt.Sprintf("%d: %s\n", offset+i+1, line))
		}
		return sb.String(), nil
	}

	// Truncate very large files
	if len(content) > 100000 {
		return content[:100000] + fmt.Sprintf("\n... (truncated, total %d bytes. Use offset/limit to read more.)", len(content)), nil
	}

	return content, nil
}

func execWriteFile(args map[string]interface{}) (string, error) {
	filePath, _ := args["file_path"].(string)
	content, _ := args["content"].(string)
	if filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	// Create parent directories
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), filePath), nil
}

func execReplace(args map[string]interface{}) (string, error) {
	filePath, _ := args["file_path"].(string)
	oldStr, _ := args["old_string"].(string)
	newStr, _ := args["new_string"].(string)
	if filePath == "" || oldStr == "" {
		return "", fmt.Errorf("file_path and old_string are required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	expectedReplacements := 1
	if e, ok := args["expected_replacements"].(float64); ok {
		expectedReplacements = int(e)
	}

	count := strings.Count(content, oldStr)
	if count == 0 {
		return "", fmt.Errorf("old_string not found in file. Make sure the text matches exactly including whitespace")
	}

	if expectedReplacements == 1 && count > 1 {
		return "", fmt.Errorf("old_string matches %d locations. Include more context to uniquely identify the target, or set expected_replacements=%d", count, count)
	}

	if count != expectedReplacements {
		return "", fmt.Errorf("expected %d replacements but found %d matches", expectedReplacements, count)
	}

	newContent := strings.Replace(content, oldStr, newStr, expectedReplacements)
	if err := os.WriteFile(filePath, []byte(newContent), 0o644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", expectedReplacements, filePath), nil
}

func execListDirectory(args map[string]interface{}) (string, error) {
	dirPath, _ := args["dir_path"].(string)
	if dirPath == "" {
		return "", fmt.Errorf("dir_path is required")
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var sb strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		sb.WriteString(name + "\n")
	}

	if sb.Len() == 0 {
		return "(empty directory)", nil
	}

	return sb.String(), nil
}

func execGlob(workDir string, args map[string]interface{}) (string, error) {
	pattern, _ := args["pattern"].(string)
	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	dir := workDir
	if d, ok := args["dir_path"].(string); ok && d != "" {
		dir = d
	}
	if dir == "" {
		dir = "."
	}

	// Use find command for recursive glob support
	cmd := exec.Command("find", dir, "-name", pattern, "-not", "-path", "*/node_modules/*", "-not", "-path", "*/.git/*", "-not", "-path", "*/vendor/*")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	// Also try shell glob for ** patterns
	if strings.Contains(pattern, "/") || strings.Contains(pattern, "**") {
		// Use bash globbing
		shellCmd := fmt.Sprintf("cd %q && find . -path %q -not -path '*/node_modules/*' -not -path '*/.git/*' 2>/dev/null | head -200", dir, pattern)
		cmd = exec.Command("bash", "-c", shellCmd)
		cmd.Stdout = &out
	}

	if err := cmd.Run(); err != nil {
		// Try simple filepath.Glob as fallback
		matches, err2 := filepath.Glob(filepath.Join(dir, pattern))
		if err2 != nil {
			return "", fmt.Errorf("glob failed: %w", err)
		}
		return strings.Join(matches, "\n"), nil
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return "(no matches found)", nil
	}

	lines := strings.Split(result, "\n")
	if len(lines) > 200 {
		lines = lines[:200]
		result = strings.Join(lines, "\n") + "\n... (truncated, 200+ matches)"
	}

	return result, nil
}

func execGrepSearch(ctx context.Context, workDir string, args map[string]interface{}) (string, error) {
	pattern, _ := args["pattern"].(string)
	if pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	dir := workDir
	if d, ok := args["dir_path"].(string); ok && d != "" {
		dir = d
	}
	if dir == "" {
		dir = "."
	}

	// Build grep command
	grepArgs := []string{"-rn", "--color=never", "-m", "100"}

	if include, ok := args["include"].(string); ok && include != "" {
		grepArgs = append(grepArgs, "--include="+include)
	}

	// Exclude common dirs
	grepArgs = append(grepArgs, "--exclude-dir=node_modules", "--exclude-dir=.git", "--exclude-dir=vendor", "--exclude-dir=dist", "--exclude-dir=build")
	grepArgs = append(grepArgs, "-E", pattern, dir)

	cmd := exec.CommandContext(ctx, "grep", grepArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	cmd.Run() // grep returns exit 1 if no matches

	result := strings.TrimSpace(out.String())
	if result == "" {
		return "(no matches found)", nil
	}

	// Limit output
	if len(result) > 50000 {
		result = result[:50000] + "\n... (output truncated)"
	}

	return result, nil
}

func execSaveMemory(workDir string, args map[string]interface{}) (string, error) {
	fact, _ := args["fact"].(string)
	if fact == "" {
		return "", fmt.Errorf("fact is required")
	}

	// Save to GEMINI.md in the working directory
	memFile := filepath.Join(workDir, "GEMINI.md")

	var existing []byte
	existing, _ = os.ReadFile(memFile)

	entry := fmt.Sprintf("\n- %s\n", fact)
	if len(existing) == 0 {
		entry = "# Project Memory\n\n- " + fact + "\n"
	}

	f, err := os.OpenFile(memFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to open memory file: %w", err)
	}
	defer f.Close()

	if len(existing) == 0 {
		if _, err := f.WriteString(entry); err != nil {
			return "", fmt.Errorf("failed to write: %w", err)
		}
	} else {
		if _, err := f.WriteString(entry); err != nil {
			return "", fmt.Errorf("failed to write: %w", err)
		}
	}

	return fmt.Sprintf("Saved to %s: %s", memFile, fact), nil
}

func execReadManyFiles(workDir string, args map[string]interface{}) (string, error) {
	includeRaw, ok := args["include"].([]interface{})
	if !ok || len(includeRaw) == 0 {
		return "", fmt.Errorf("include patterns are required")
	}

	var excludePatterns []string
	if excludeRaw, ok := args["exclude"].([]interface{}); ok {
		for _, e := range excludeRaw {
			if s, ok := e.(string); ok {
				excludePatterns = append(excludePatterns, s)
			}
		}
	}

	dir := workDir
	if dir == "" {
		dir = "."
	}

	var sb strings.Builder
	totalSize := 0
	fileCount := 0
	const maxTotalSize = 500000
	const maxFiles = 100

	for _, pat := range includeRaw {
		pattern, ok := pat.(string)
		if !ok {
			continue
		}

		// Use find + glob
		shellCmd := fmt.Sprintf("cd %q && find . -path %q -not -path '*/node_modules/*' -not -path '*/.git/*' -not -path '*/vendor/*' -type f 2>/dev/null | head -200",
			dir, pattern)
		cmd := exec.Command("bash", "-c", shellCmd)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Run()

		matches := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(matches) == 1 && matches[0] == "" {
			// Try filepath.Glob as fallback
			globMatches, _ := filepath.Glob(filepath.Join(dir, pattern))
			matches = nil
			for _, m := range globMatches {
				matches = append(matches, m)
			}
		}

		for _, match := range matches {
			if match == "" || fileCount >= maxFiles || totalSize >= maxTotalSize {
				continue
			}

			// Resolve to absolute path
			fullPath := match
			if !filepath.IsAbs(match) {
				fullPath = filepath.Join(dir, match)
			}

			// Check exclude patterns
			excluded := false
			for _, ep := range excludePatterns {
				if matched, _ := filepath.Match(ep, filepath.Base(fullPath)); matched {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}

			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			content := string(data)
			if totalSize+len(content) > maxTotalSize {
				content = content[:maxTotalSize-totalSize] + "\n... (truncated)"
			}

			sb.WriteString(fmt.Sprintf("=== %s ===\n", fullPath))
			sb.WriteString(content)
			sb.WriteString("\n\n")
			totalSize += len(content)
			fileCount++
		}
	}

	if fileCount == 0 {
		return "(no files matched the patterns)", nil
	}

	return fmt.Sprintf("Read %d files:\n\n%s", fileCount, sb.String()), nil
}

func execGoogleWebSearch(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	// Use curl + HTML parsing via a simple approach
	escapedQuery := strings.ReplaceAll(query, " ", "+")
	escapedQuery = strings.ReplaceAll(escapedQuery, "&", "%26")

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use ddgr (DuckDuckGo) or curl Google
	cmd := exec.CommandContext(timeoutCtx, "bash", "-c",
		fmt.Sprintf(`curl -sL "https://html.duckduckgo.com/html/?q=%s" 2>/dev/null | grep -oP '(?<=<a rel="nofollow" class="result__a" href=")[^"]+' | head -10`, escapedQuery))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	cmd.Run()

	result := strings.TrimSpace(out.String())
	if result == "" {
		// Fallback: use search via lynx or w3m if available
		cmd2 := exec.CommandContext(timeoutCtx, "bash", "-c",
			fmt.Sprintf(`curl -sL -A "Mozilla/5.0" "https://www.google.com/search?q=%s&num=10" 2>/dev/null | sed 's/<[^>]*>//g' | grep -v "^$" | head -50`, escapedQuery))
		var out2 bytes.Buffer
		cmd2.Stdout = &out2
		cmd2.Run()
		result = strings.TrimSpace(out2.String())
	}

	if result == "" {
		return "No search results found. Try a different query.", nil
	}

	if len(result) > 20000 {
		result = result[:20000] + "\n... (truncated)"
	}

	return fmt.Sprintf("Search results for: %s\n\n%s", query, result), nil
}

func execWebFetch(ctx context.Context, args map[string]interface{}) (string, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return "", fmt.Errorf("url is required")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("url must start with http:// or https://")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Fetch URL and convert to plain text
	cmd := exec.CommandContext(timeoutCtx, "bash", "-c",
		fmt.Sprintf(`curl -sL -A "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)" --max-time 25 %q 2>/dev/null | sed 's/<script[^>]*>.*<\/script>//g; s/<style[^>]*>.*<\/style>//g; s/<[^>]*>//g; s/&nbsp;/ /g; s/&amp;/\&/g; s/&lt;/</g; s/&gt;/>/g; s/&quot;/"/g' | sed '/^$/d' | head -500`, url))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return "(empty response from URL)", nil
	}

	if len(result) > 100000 {
		result = result[:100000] + "\n... (truncated)"
	}

	return fmt.Sprintf("Content from %s:\n\n%s", url, result), nil
}

func execActivateSkill(workDir string, args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)

	// Search skill directories: project-local first, then global
	homeDir, _ := os.UserHomeDir()
	skillDirs := []string{
		filepath.Join(workDir, ".gmn", "skills"),
		filepath.Join(homeDir, ".gmn", "skills"),
	}

	if name == "" {
		// List available skills
		var skills []string
		for _, dir := range skillDirs {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					skills = append(skills, e.Name())
				}
			}
		}
		if len(skills) == 0 {
			return "No skills found. Create skills in .gmn/skills/<name>/instruction.md", nil
		}
		return fmt.Sprintf("Available skills: %s", strings.Join(skills, ", ")), nil
	}

	// Find and activate the named skill
	for _, dir := range skillDirs {
		instrPath := filepath.Join(dir, name, "instruction.md")
		data, err := os.ReadFile(instrPath)
		if err != nil {
			continue
		}
		return fmt.Sprintf("<activated_skill name=%q>\n%s\n</activated_skill>", name, string(data)), nil
	}

	// List what's available for a helpful error
	var available []string
	for _, dir := range skillDirs {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if e.IsDir() {
				available = append(available, e.Name())
			}
		}
	}
	if len(available) > 0 {
		return "", fmt.Errorf("skill %q not found. Available: %s", name, strings.Join(available, ", "))
	}
	return "", fmt.Errorf("skill %q not found and no skills directory exists", name)
}

func execGetInternalDocs(workDir string, args map[string]interface{}) (string, error) {
	docPath, _ := args["path"].(string)

	docsDir := filepath.Join(workDir, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return "No docs/ directory found in the project.", nil
	}

	if docPath == "" {
		// List all docs
		var docs []string
		filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				rel, _ := filepath.Rel(docsDir, path)
				docs = append(docs, rel)
			}
			return nil
		})
		if len(docs) == 0 {
			return "No documentation files found in docs/.", nil
		}
		return fmt.Sprintf("Available documentation (%d files):\n%s", len(docs), strings.Join(docs, "\n")), nil
	}

	// Read specific doc
	fullPath := filepath.Join(docsDir, docPath)
	// Security: prevent path traversal
	absDocsDir, _ := filepath.Abs(docsDir)
	absFullPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFullPath, absDocsDir) {
		return "", fmt.Errorf("path traversal not allowed")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read doc: %w", err)
	}
	return string(data), nil
}

func execWriteTodos(workDir string, args map[string]interface{}) (string, error) {
	todosRaw, ok := args["todos"].([]interface{})
	if !ok {
		return "", fmt.Errorf("todos array is required")
	}

	type TodoItem struct {
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	var todos []TodoItem
	for _, t := range todosRaw {
		m, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		desc, _ := m["description"].(string)
		status, _ := m["status"].(string)
		if desc != "" {
			todos = append(todos, TodoItem{Description: desc, Status: status})
		}
	}

	// Format as markdown checklist
	var sb strings.Builder
	sb.WriteString("# Todo List\n\n")
	pending, inProgress, completed, cancelled := 0, 0, 0, 0
	for _, todo := range todos {
		switch todo.Status {
		case "completed":
			sb.WriteString(fmt.Sprintf("- [x] %s\n", todo.Description))
			completed++
		case "cancelled":
			sb.WriteString(fmt.Sprintf("- [~] ~~%s~~\n", todo.Description))
			cancelled++
		case "in_progress":
			sb.WriteString(fmt.Sprintf("- [ ] 🔄 %s\n", todo.Description))
			inProgress++
		default:
			sb.WriteString(fmt.Sprintf("- [ ] %s\n", todo.Description))
			pending++
		}
	}

	// Save to .todos.md in work dir
	todoFile := filepath.Join(workDir, ".todos.md")
	if err := os.WriteFile(todoFile, []byte(sb.String()), 0o644); err != nil {
		return "", fmt.Errorf("failed to write todos: %w", err)
	}

	return fmt.Sprintf("Updated %d todos (pending: %d, in_progress: %d, completed: %d, cancelled: %d)\n\n%s",
		len(todos), pending, inProgress, completed, cancelled, sb.String()), nil
}
