package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/dutakey/shellgate/internal/types"
)

type CodexExecutor struct {
	binary     string
	sandbox    string
	workingDir string
}

func NewCodexExecutor(binary, sandbox, workingDir string) *CodexExecutor {
	if binary == "" {
		binary = "codex"
	}
	return &CodexExecutor{
		binary:     binary,
		sandbox:    sandbox,
		workingDir: workingDir,
	}
}

// Stream spawns `codex exec --json <prompt>` and emits text chunks.
// Channel closed when process exits or ctx cancelled.
func (e *CodexExecutor) Stream(ctx context.Context, prompt, model, reasoningEffort string) (<-chan string, <-chan error) {
	events := make(chan string, 32)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		cmd := e.buildCmd(ctx, prompt, model, reasoningEffort)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errc <- fmt.Errorf("stdout pipe: %w", err)
			return
		}
		cmd.Stderr = io.Discard

		if err := cmd.Start(); err != nil {
			errc <- fmt.Errorf("codex start: %w", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var base types.CodexEvent
			if err := json.Unmarshal([]byte(line), &base); err != nil {
				continue
			}

			switch base.Type {
			case "item.completed":
				var ev types.CodexItemEvent
				if err := json.Unmarshal([]byte(line), &ev); err == nil {
					if ev.Item.Type == "agent_message" && ev.Item.Text != "" {
						select {
						case events <- ev.Item.Text:
						case <-ctx.Done():
							cmd.Process.Kill()
							return
						}
					}
				}

			case "turn.failed", "error":
				var ev types.CodexErrorEvent
				if err := json.Unmarshal([]byte(line), &ev); err == nil {
					errc <- fmt.Errorf("codex error: %s", ev.ErrorMessage())
					cmd.Process.Kill()
					return
				}
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			errc <- fmt.Errorf("read stdout: %w", err)
			return
		}

		if err := cmd.Wait(); err != nil && ctx.Err() == nil {
			errc <- fmt.Errorf("codex exit: %w", err)
		}
	}()

	return events, errc
}

// Exec runs codex exec and collects full response (non-streaming)
func (e *CodexExecutor) Exec(ctx context.Context, prompt, model, reasoningEffort string) (string, error) {
	events, errc := e.Stream(ctx, prompt, model, reasoningEffort)

	var parts []string
	for text := range events {
		parts = append(parts, text)
	}

	if err := <-errc; err != nil {
		return "", err
	}

	return strings.Join(parts, ""), nil
}

func (e *CodexExecutor) buildCmd(ctx context.Context, prompt, model, reasoningEffort string) *exec.Cmd {
	args := []string{"exec", "--json", "--ephemeral", "--skip-git-repo-check"}

	if model != "" {
		args = append(args, "--model", model)
	}

	if reasoningEffort != "" {
		args = append(args, "-c", "model_reasoning_effort="+reasoningEffort)
	}

	if e.sandbox != "" && e.sandbox != "read-only" {
		args = append(args, "--sandbox", e.sandbox)
	}

	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Env = inheritEnv()

	if e.workingDir != "" {
		cmd.Dir = e.workingDir
	} else {
		cmd.Dir = os.TempDir()
	}

	return cmd
}
