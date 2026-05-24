package executor

import (
	"context"
	"strings"
)

// Route maps a model prefix to an Executor.
type Route struct {
	Prefix   string
	Executor Executor
}

// RouterExecutor routes requests to the appropriate backend based on model prefix.
// Falls back to defaultExec when no prefix matches.
type RouterExecutor struct {
	routes      []Route
	defaultExec Executor
}

func NewRouterExecutor(defaultExec Executor, routes ...Route) *RouterExecutor {
	return &RouterExecutor{
		routes:      routes,
		defaultExec: defaultExec,
	}
}

func (r *RouterExecutor) resolve(model string) Executor {
	lower := strings.ToLower(model)
	for _, rt := range r.routes {
		if strings.HasPrefix(lower, rt.Prefix) {
			return rt.Executor
		}
	}
	return r.defaultExec
}

func (r *RouterExecutor) Stream(ctx context.Context, prompt, model, reasoningEffort string) (<-chan string, <-chan error) {
	return r.resolve(model).Stream(ctx, prompt, model, reasoningEffort)
}

func (r *RouterExecutor) Exec(ctx context.Context, prompt, model, reasoningEffort string) (string, error) {
	return r.resolve(model).Exec(ctx, prompt, model, reasoningEffort)
}
