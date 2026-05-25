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
)

type ClaudeExecutor struct {
	binary     string
	workingDir string
}

func NewClaudeExecutor(binary, workingDir string) *ClaudeExecutor {
	if binary == "" {
		binary = "claude"
	}
	return &ClaudeExecutor{binary: binary, workingDir: workingDir}
}

// Claude Code CLI stream-json event types
type claudeLineEvent struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype"`
	Event   *claudeAPIEvent `json:"event,omitempty"`
	Result  string          `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type claudeAPIEvent struct {
	Type  string        `json:"type"`
	Delta *claudeDelta  `json:"delta,omitempty"`
}

type claudeDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (e *ClaudeExecutor) Stream(ctx context.Context, prompt, model, reasoningEffort string) (<-chan string, <-chan error) {
	events := make(chan string, 32)
	errc := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errc)

		cmd := e.buildCmd(ctx, prompt, model, reasoningEffort)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			errc <- fmt.Errorf("claude stdout pipe: %w", err)
			return
		}
		cmd.Stderr = io.Discard

		if err := cmd.Start(); err != nil {
			errc <- fmt.Errorf("claude start: %w", err)
			return
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var ev claudeLineEvent
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				continue
			}

			switch ev.Type {
			case "stream_event":
				if ev.Event != nil && ev.Event.Delta != nil &&
					ev.Event.Delta.Type == "text_delta" && ev.Event.Delta.Text != "" {
					select {
					case events <- ev.Event.Delta.Text:
					case <-ctx.Done():
						cmd.Process.Kill()
						return
					}
				}
			case "result":
				if ev.Subtype == "error" {
					errc <- fmt.Errorf("claude error: %s", ev.Error)
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
			errc <- fmt.Errorf("claude exit: %w", err)
		}
	}()

	return events, errc
}

func (e *ClaudeExecutor) Exec(ctx context.Context, prompt, model, reasoningEffort string) (string, error) {
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

func (e *ClaudeExecutor) buildCmd(ctx context.Context, prompt, model, reasoningEffort string) *exec.Cmd {
	args := []string{
		"-p", prompt,
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages",
	}

	if model != "" {
		args = append(args, "--model", model)
	}

	switch reasoningEffort {
	case "medium", "high", "extra-high":
		args = append(args, "--think")
	}

	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Env = inheritEnv()

	if e.workingDir != "" {
		cmd.Dir = e.workingDir
	} else {
		// run from /tmp — no CLAUDE.md, no project memory, no git context
		cmd.Dir = os.TempDir()
	}

	return cmd
}
