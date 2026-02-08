package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestBuiltinTools(t *testing.T) {
	ctx := context.Background()
	workDir, _ := os.Getwd()

	fmt.Println("\n=== Testing Builtin Tools on Windows ===")

	// Test 1: Shell command execution
	t.Run("ShellCommand", func(t *testing.T) {
		fmt.Println("1. Testing shell command (go version)...")
		args := map[string]interface{}{
			"command": "go version",
			"timeout": 10000,
		}
		result, err := execShellCommand(ctx, workDir, args)
		if err != nil {
			t.Errorf("❌ Error: %v", err)
		} else {
			fmt.Printf("   ✅ Success: %s\n", result[:min(len(result), 100)])
		}
	})

	// Test 2: Glob file search
	t.Run("Glob", func(t *testing.T) {
		fmt.Println("\n2. Testing glob (find .go files in service/)...")
		args := map[string]interface{}{
			"pattern": "service/*.go",
		}
		result, err := execGlob(workDir, args)
		if err != nil {
			t.Errorf("❌ Error: %v", err)
		} else {
			lines := 0
			fmt.Println("   ✅ Success: Found files (first 5 lines):")
			for i, c := range result {
				if c == '\n' {
					lines++
					if lines >= 5 {
						fmt.Println("   ...")
						break
					}
				}
				if i < 500 {
					fmt.Print(string(c))
				}
			}
		}
	})

	// Test 3: Grep search
	t.Run("GrepSearch", func(t *testing.T) {
		fmt.Println("\n3. Testing grep (search for 'runtime.GOOS')...")
		args := map[string]interface{}{
			"pattern": "runtime\\.GOOS",
			"glob":    "*.go",
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		result, err := execGrepSearch(ctx, workDir, args)
		if err != nil {
			t.Errorf("❌ Error: %v", err)
		} else {
			lines := 0
			fmt.Println("   ✅ Success: Found matches:")
			for i, c := range result {
				if c == '\n' {
					lines++
					if lines >= 5 {
						fmt.Println("   ...")
						break
					}
				}
				if i < 500 {
					fmt.Print(string(c))
				}
			}
		}
	})

	// Test 4: Web fetch (requires API auth, skip in CI)
	t.Run("WebFetch", func(t *testing.T) {
		t.Skip("Skipping: requires SettingsService with API auth")
	})

	// Test 5: Web search (requires API auth, skip in CI)
	t.Run("WebSearchManila", func(t *testing.T) {
		t.Skip("Skipping: requires SettingsService with API auth")
	})

	fmt.Println("\n=== All tests completed ===")
}
