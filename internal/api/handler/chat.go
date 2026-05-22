package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/dutakey/shellgate/internal/executor"
	"github.com/dutakey/shellgate/internal/formatter"
	"github.com/dutakey/shellgate/internal/types"
)

type ChatHandler struct {
	exec   *executor.CodexExecutor
	logger *zap.Logger
}

func NewChatHandler(exec *executor.CodexExecutor, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{exec: exec, logger: logger}
}

func (h *ChatHandler) ChatCompletions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req types.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body", "invalid_request_error")
			return
		}

		if len(req.Messages) == 0 {
			writeJSONError(w, http.StatusBadRequest, "messages cannot be empty", "invalid_request_error")
			return
		}

		if req.Model == "" {
			req.Model = "codex"
		}

		prompt := formatter.BuildPrompt(req.Messages)

		h.logger.Info("chat request",
			zap.String("model", req.Model),
			zap.Bool("stream", req.Stream),
			zap.Int("messages", len(req.Messages)),
		)

		if req.Stream {
			h.handleStream(w, r, req.Model, prompt)
		} else {
			h.handleSync(w, r, req.Model, prompt)
		}
	}
}

func (h *ChatHandler) handleSync(w http.ResponseWriter, r *http.Request, model, prompt string) {
	content, err := h.exec.Exec(r.Context(), prompt)
	if err != nil {
		h.logger.Error("codex exec failed", zap.Error(err))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("codex exec failed: %s", err.Error()), "upstream_error")
		return
	}

	id := formatter.NewCompletionID()
	resp := formatter.BuildResponse(id, model, content)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ChatHandler) handleStream(w http.ResponseWriter, r *http.Request, model, prompt string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming not supported", "internal_error")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	id := formatter.NewCompletionID()

	sendChunk := func(chunk interface{}) {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	sendChunk(formatter.BuildFirstChunk(id, model))

	events, errc := h.exec.Stream(r.Context(), prompt)

	for ev := range events {
		if ev.Item.Text != "" {
			sendChunk(formatter.BuildChunk(id, model, ev.Item.Text, nil))
		}
	}

	if err := <-errc; err != nil {
		h.logger.Error("codex stream error", zap.Error(err))
		errData, _ := json.Marshal(types.ErrorResponse{
			Error: types.APIError{
				Message: err.Error(),
				Type:    "upstream_error",
			},
		})
		fmt.Fprintf(w, "data: %s\n\n", errData)
		flusher.Flush()
		return
	}

	sendChunk(formatter.BuildStopChunk(id, model))
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func writeJSONError(w http.ResponseWriter, status int, msg, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(types.ErrorResponse{
		Error: types.APIError{
			Message: msg,
			Type:    errType,
		},
	})
}
