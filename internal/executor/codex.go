package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// Stream spawns `codex exec --json <prompt>` and emits CodexItemEvents.
// Channel closed when process exits or ctx cancelled.
func (e *CodexExecutor) Stream(ctx context.Context, prompt string) (<-chan *types.CodexItemEvent, <-chan error) {
	events := make(chan *types.CodexItemEvent, 32)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		cmd := e.buildCmd(ctx, prompt)
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
					if ev.Item.Type == "agent_message" {
						select {
						case events <- &ev:
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
func (e *CodexExecutor) Exec(ctx context.Context, prompt string) (string, error) {
	events, errc := e.Stream(ctx, prompt)

	var parts []string
	for ev := range events {
		parts = append(parts, ev.Item.Text)
	}

	if err := <-errc; err != nil {
		return "", err
	}

	return strings.Join(parts, ""), nil
}

func (e *CodexExecutor) buildCmd(ctx context.Context, prompt string) *exec.Cmd {
	args := []string{"exec", "--json", "--ephemeral", "--skip-git-repo-check"}

	if e.sandbox != "" && e.sandbox != "read-only" {
		args = append(args, "--sandbox", e.sandbox)
	}

	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Env = inheritEnv()

	if e.workingDir != "" {
		cmd.Dir = e.workingDir
	}

	return cmd
}
