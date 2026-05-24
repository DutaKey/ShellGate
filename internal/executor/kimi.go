package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type KimiExecutor struct {
	binary     string
	workingDir string
}

func NewKimiExecutor(binary, workingDir string) *KimiExecutor {
	if binary == "" {
		binary = "kimi"
	}
	return &KimiExecutor{binary: binary, workingDir: workingDir}
}

type kimiOutput struct {
	Role    string        `json:"role"`
	Content []kimiContent `json:"content"`
}

type kimiContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (e *KimiExecutor) Stream(ctx context.Context, prompt, model, reasoningEffort string) (<-chan string, <-chan error) {
	out := make(chan string, 1)
	errc := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errc)

		text, err := e.run(ctx, prompt, model, reasoningEffort)
		if err != nil {
			errc <- err
			return
		}
		if text != "" {
			out <- text
		}
	}()

	return out, errc
}

func (e *KimiExecutor) Exec(ctx context.Context, prompt, model, reasoningEffort string) (string, error) {
	return e.run(ctx, prompt, model, reasoningEffort)
}

func (e *KimiExecutor) run(ctx context.Context, prompt, model, reasoningEffort string) (string, error) {
	cmd := e.buildCmd(ctx, prompt, model, reasoningEffort)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("kimi stdout pipe: %w", err)
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("kimi start: %w", err)
	}

	raw, err := io.ReadAll(stdout)
	if err != nil {
		_ = cmd.Process.Kill()
		return "", fmt.Errorf("kimi read stdout: %w", err)
	}

	if err := cmd.Wait(); err != nil && ctx.Err() == nil {
		return "", fmt.Errorf("kimi exit: %w", err)
	}

	return parseKimiOutput(raw)
}

func (e *KimiExecutor) buildCmd(ctx context.Context, prompt, model, reasoningEffort string) *exec.Cmd {
	args := []string{"--print", "--output-format", "stream-json", "--yolo", "-p", prompt}

	if model != "" {
		args = append(args, "-m", model)
	}

	switch reasoningEffort {
	case "none", "low":
		args = append(args, "--no-thinking")
	case "medium", "high", "extra-high":
		args = append(args, "--thinking")
	}

	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Env = inheritEnv()

	if e.workingDir != "" {
		cmd.Dir = e.workingDir
	}

	return cmd
}

func parseKimiOutput(raw []byte) (string, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return "", nil
	}

	var output kimiOutput
	if err := json.Unmarshal([]byte(trimmed), &output); err != nil {
		return "", fmt.Errorf("kimi parse output: %w", err)
	}

	var parts []string
	for _, c := range output.Content {
		if c.Type == "text" && c.Text != "" {
			parts = append(parts, c.Text)
		}
	}

	return strings.Join(parts, ""), nil
}
