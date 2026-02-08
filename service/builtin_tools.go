package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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
func ExecuteBuiltinTool(ctx context.Context, workDir string, name string, args map[string]interface{}, settings *SettingsService) (string, error) {
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
		return execGoogleWebSearch(ctx, args, settings)
	case "web_fetch":
		return execWebFetch(ctx, args, settings)
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

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: use cmd /c for simple commands, powershell for complex ones
		cmd = exec.CommandContext(timeoutCtx, "powershell", "-NoProfile", "-Command", command)
	} else {
		// Unix-like: use bash
		cmd = exec.CommandContext(timeoutCtx, "bash", "-c", command)
	}

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

	// Use cross-platform Go implementation instead of shell commands
	var matches []string
	excludeDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
		".vscode":      true,
	}

	// Handle ** patterns
	if strings.Contains(pattern, "**") {
		// Recursive glob: walk directory tree
		parts := strings.Split(pattern, "**")
		baseName := ""
		if len(parts) > 1 {
			baseName = strings.TrimPrefix(strings.TrimPrefix(parts[1], "/"), "\\")
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			// Skip excluded directories
			if info.IsDir() {
				if excludeDirs[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}

			// Match pattern
			if baseName == "" || strings.HasSuffix(path, baseName) {
				relPath, _ := filepath.Rel(dir, path)
				matches = append(matches, filepath.ToSlash(relPath))
			}
			if len(matches) >= 200 {
				return filepath.SkipAll
			}
			return nil
		})
		if err != nil && err != filepath.SkipAll {
			return "", fmt.Errorf("walk failed: %w", err)
		}
	} else {
		// Simple glob without **
		fullPattern := filepath.Join(dir, pattern)
		m, err := filepath.Glob(fullPattern)
		if err != nil {
			return "", fmt.Errorf("glob failed: %w", err)
		}
		for _, match := range m {
			relPath, _ := filepath.Rel(dir, match)
			matches = append(matches, filepath.ToSlash(relPath))
		}
	}

	if len(matches) == 0 {
		return "(no matches found)", nil
	}

	if len(matches) > 200 {
		matches = matches[:200]
		return strings.Join(matches, "\n") + "\n... (truncated, 200+ matches)", nil
	}

	return strings.Join(matches, "\n"), nil
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

	includePattern, _ := args["include"].(string)

	// Compile regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Cross-platform grep implementation
	excludeDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
	}

	var results []string
	matchCount := 0
	maxMatches := 500

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip excluded directories
		if info.IsDir() {
			if excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file pattern filter
		if includePattern != "" {
			matched, _ := filepath.Match(includePattern, info.Name())
			if !matched {
				return nil
			}
		}

		// Skip binary files (simple heuristic)
		if isBinaryFile(path) {
			return nil
		}

		// Search in file
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		relPath, _ := filepath.Rel(dir, path)
		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d:%s", filepath.ToSlash(relPath), lineNum, line))
				matchCount++

				if matchCount >= maxMatches {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return "(no matches found)", nil
	}

	result := strings.Join(results, "\n")
	if len(result) > 50000 {
		result = result[:50000] + "\n... (output truncated)"
	}

	return result, nil
}

// isBinaryFile checks if a file is likely binary (simple heuristic)
func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := map[string]bool{
		".txt": true, ".md": true, ".go": true, ".js": true, ".ts": true,
		".py": true, ".java": true, ".c": true, ".cpp": true, ".h": true,
		".json": true, ".xml": true, ".yaml": true, ".yml": true, ".toml": true,
		".html": true, ".css": true, ".scss": true, ".vue": true, ".jsx": true,
		".tsx": true, ".rs": true, ".rb": true, ".php": true, ".sh": true,
		".sql": true, ".proto": true, ".graphql": true, ".env": true,
		".conf": true, ".cfg": true, ".ini": true, ".log": true,
	}
	return !textExts[ext] && ext != ""
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

func execGoogleWebSearch(ctx context.Context, args map[string]interface{}, settings *SettingsService) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	// Get API client
	client, err := settings.EnsureAuth(ctx)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Build request with Google Search Grounding
	req := &api.GenerateRequest{
		Model:   settings.GetDefaultModel(),
		Project: settings.GetProjectID(),
		Request: api.InnerRequest{
			Contents: []api.Content{
				{
					Role: "user",
					Parts: []api.Part{
						{Text: query},
					},
				},
			},
			Tools: []api.Tool{
				{
					GoogleSearchRetrieval: &api.GoogleSearchRetrieval{
						DynamicRetrievalConfig: &api.DynamicRetrievalConfig{
							Mode:             "MODE_DYNAMIC",
							DynamicThreshold: 0.3,
						},
					},
				},
			},
		},
	}

	// Send request to Gemini API with Google Search Grounding
	resp, err := client.Generate(timeoutCtx, req)
	if err != nil {
		return "", fmt.Errorf("Google Search failed: %w", err)
	}

	// Extract text from response
	if len(resp.Response.Candidates) == 0 {
		return "No search results found.", nil
	}

	var resultText strings.Builder
	for _, part := range resp.Response.Candidates[0].Content.Parts {
		if part.Text != "" {
			resultText.WriteString(part.Text)
		}
	}

	if resultText.Len() == 0 {
		return "No search results found.", nil
	}

	return resultText.String(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func execWebFetch(ctx context.Context, args map[string]interface{}, settings *SettingsService) (string, error) {
	urlStr, _ := args["url"].(string)
	if urlStr == "" {
		return "", fmt.Errorf("url is required")
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "", fmt.Errorf("url must start with http:// or https://")
	}

	// Get prompt parameter (optional - for specific instructions)
	prompt, _ := args["prompt"].(string)
	if prompt == "" {
		prompt = fmt.Sprintf("Please summarize the content from %s", urlStr)
	} else if !strings.Contains(prompt, urlStr) {
		prompt = fmt.Sprintf("%s\n\nURL: %s", prompt, urlStr)
	}

	// Try using Gemini API with Grounding first
	client, err := settings.EnsureAuth(ctx)
	if err == nil {
		timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		req := &api.GenerateRequest{
			Model:   settings.GetDefaultModel(),
			Project: settings.GetProjectID(),
			Request: api.InnerRequest{
				Contents: []api.Content{
					{
						Role: "user",
						Parts: []api.Part{
							{Text: prompt},
						},
					},
				},
				Tools: []api.Tool{
					{
						GoogleSearchRetrieval: &api.GoogleSearchRetrieval{
							DynamicRetrievalConfig: &api.DynamicRetrievalConfig{
								Mode:             "MODE_DYNAMIC",
								DynamicThreshold: 0.3,
							},
						},
					},
				},
			},
		}

		resp, err := client.Generate(timeoutCtx, req)
		if err == nil && len(resp.Response.Candidates) > 0 {
			var resultText strings.Builder
			for _, part := range resp.Response.Candidates[0].Content.Parts {
				if part.Text != "" {
					resultText.WriteString(part.Text)
				}
			}
			if resultText.Len() > 0 {
				return resultText.String(), nil
			}
		}
		// If API method fails, fall back to local fetch
	}

	// Fallback: Local HTTP fetch
	return execWebFetchLocal(ctx, urlStr)
}

func execWebFetchLocal(ctx context.Context, urlStr string) (string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Cross-platform HTTP request
	req, err := http.NewRequestWithContext(timeoutCtx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	httpClient := &http.Client{
		Timeout: 25 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 500*1024)) // Limit to 500KB
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Simple HTML to text conversion
	text := stripHTMLTags(string(body))
	text = htmlDecode(text)
	text = strings.TrimSpace(text)

	if text == "" {
		return "(empty response from URL)", nil
	}

	if len(text) > 100000 {
		text = text[:100000] + "\n... (truncated)"
	}

	return fmt.Sprintf("Content from %s:\n\n%s", urlStr, text), nil
}

func stripHTMLTags(html string) string {
	// Remove script and style tags
	scriptRe := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = scriptRe.ReplaceAllString(html, "")

	styleRe := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = styleRe.ReplaceAllString(html, "")

	// Remove all HTML tags
	tagRe := regexp.MustCompile(`<[^>]*>`)
	text := tagRe.ReplaceAllString(html, " ")

	// Remove extra whitespace
	spaceRe := regexp.MustCompile(`\s+`)
	text = spaceRe.ReplaceAllString(text, " ")

	return text
}

func htmlDecode(s string) string {
	replacements := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  "\"",
		"&#39;":   "'",
		"&apos;":  "'",
		"&ndash;": "-",
		"&mdash;": "—",
	}

	for entity, char := range replacements {
		s = strings.ReplaceAll(s, entity, char)
	}
	return s
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
