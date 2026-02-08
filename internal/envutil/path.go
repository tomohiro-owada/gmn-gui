// Package envutil provides environment helpers for GUI apps.
// macOS GUI apps don't inherit the user's shell PATH, so tools like
// node, npx, etc. installed via Homebrew, nvm, or volta won't be found.
// On Windows, the PATH is typically inherited correctly, so no extra paths are added.
package envutil

import (
	"os"
	"runtime"
	"strings"
)

// ShellEnv returns os.Environ() with extra PATH entries for common tool locations.
// On Windows, it returns os.Environ() unmodified since GUI apps inherit PATH correctly.
func ShellEnv() []string {
	if runtime.GOOS == "windows" {
		return os.Environ()
	}

	home := os.Getenv("HOME")
	extraPaths := []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		home + "/.volta/bin",
		home + "/.local/bin",
		home + "/.bun/bin",
	}

	// Try to find the active nvm node version
	if entries, err := os.ReadDir(home + "/.nvm/versions/node"); err == nil {
		for i := len(entries) - 1; i >= 0; i-- {
			if entries[i].IsDir() {
				extraPaths = append(extraPaths, home+"/.nvm/versions/node/"+entries[i].Name()+"/bin")
				break
			}
		}
	}

	env := os.Environ()
	sep := string(os.PathListSeparator)
	pathPrefix := strings.Join(extraPaths, sep)

	found := false
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + pathPrefix + sep + e[5:]
			found = true
			break
		}
	}
	if !found {
		env = append(env, "PATH="+pathPrefix+sep+"/usr/bin:/bin")
	}

	return env
}
