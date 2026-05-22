package executor

import "os"

// inheritEnv returns current process env.
// codex CLI reads saved OAuth credentials from $CODEX_HOME (~/.codex by default).
func inheritEnv() []string {
	return os.Environ()
}
