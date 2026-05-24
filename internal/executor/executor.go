package executor

import "context"

// Executor is the provider-agnostic interface for running prompts through a CLI backend.
// Stream emits zero or more text chunks, then closes both channels.
// Exec is a convenience wrapper that collects all chunks into a single string.
type Executor interface {
	Stream(ctx context.Context, prompt, model, reasoningEffort string) (<-chan string, <-chan error)
	Exec(ctx context.Context, prompt, model, reasoningEffort string) (string, error)
}
