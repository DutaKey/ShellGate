package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"

	"github.com/dutakey/shellgate/internal/executor"
	"github.com/dutakey/shellgate/internal/formatter"
	"github.com/dutakey/shellgate/internal/types"
)

type ResponsesHandler struct {
	exec   executor.Executor
	logger *zap.Logger
}

func NewResponsesHandler(exec executor.Executor, logger *zap.Logger) *ResponsesHandler {
	return &ResponsesHandler{exec: exec, logger: logger}
}

func (h *ResponsesHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req types.ResponsesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body", "invalid_request_error")
			return
		}

		if req.Model == "" {
			req.Model = "gpt-5.4"
		}

		prompt := extractPrompt(req.Input)
		if prompt == "" {
			writeJSONError(w, http.StatusBadRequest, "input cannot be empty", "invalid_request_error")
			return
		}

		h.logger.Info("responses request",
			zap.String("model", req.Model),
			zap.Bool("stream", req.Stream),
		)

		if req.Stream {
			h.handleStream(w, r, req.Model, prompt, req.ReasoningEffort)
		} else {
			h.handleSync(w, r, req.Model, prompt, req.ReasoningEffort)
		}
	}
}

func (h *ResponsesHandler) handleSync(w http.ResponseWriter, r *http.Request, model, prompt, reasoningEffort string) {
	content, err := h.exec.Exec(r.Context(), prompt, model, reasoningEffort)
	if err != nil {
		h.logger.Error("exec failed", zap.Error(err))
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("exec failed: %s", err.Error()), "upstream_error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildResponsesResponse(model, content))
}

func (h *ResponsesHandler) handleStream(w http.ResponseWriter, r *http.Request, model, prompt, reasoningEffort string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming not supported", "internal_error")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	respID := "resp-" + uuid.New().String()
	msgID := "msg-" + uuid.New().String()

	send := func(eventType string, data interface{}) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, b)
		flusher.Flush()
	}

	send("response.created", map[string]interface{}{
		"type": "response.created",
		"response": map[string]interface{}{
			"id": respID, "object": "response", "status": "in_progress",
			"model": model, "output": []interface{}{}, "tools": []interface{}{},
		},
	})

	send("response.output_item.added", map[string]interface{}{
		"type": "response.output_item.added", "output_index": 0,
		"item": map[string]interface{}{
			"id": msgID, "type": "message", "status": "in_progress",
			"role": "assistant", "content": []interface{}{},
		},
	})

	send("response.content_part.added", map[string]interface{}{
		"type": "response.content_part.added",
		"item_id": msgID, "output_index": 0, "content_index": 0,
		"part": map[string]interface{}{"type": "output_text", "text": "", "annotations": []interface{}{}},
	})

	events, errc := h.exec.Stream(r.Context(), prompt, model, reasoningEffort)

	for text := range events {
		if text != "" {
			send("response.output_text.delta", map[string]interface{}{
				"type":    "response.output_text.delta",
				"item_id": msgID, "output_index": 0, "content_index": 0,
				"delta": text,
			})
		}
	}

	if err := <-errc; err != nil {
		h.logger.Error("stream error", zap.Error(err))
		send("error", map[string]interface{}{
			"type": "error", "code": "upstream_error", "message": err.Error(),
		})
		return
	}

	send("response.output_text.done", map[string]interface{}{
		"type":    "response.output_text.done",
		"item_id": msgID, "output_index": 0, "content_index": 0,
	})
	send("response.output_item.done", map[string]interface{}{
		"type": "response.output_item.done", "output_index": 0,
		"item": map[string]interface{}{
			"id": msgID, "type": "message", "status": "completed",
			"role": "assistant", "content": []interface{}{
				map[string]interface{}{"type": "output_text", "annotations": []interface{}{}},
			},
		},
	})
	send("response.completed", map[string]interface{}{
		"type": "response.completed",
		"response": map[string]interface{}{
			"id": respID, "object": "response", "status": "completed",
			"model": model, "tools": []interface{}{},
		},
	})
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func buildResponsesResponse(model, content string) *types.ResponsesResponse {
	return &types.ResponsesResponse{
		ID:                "resp-" + uuid.New().String(),
		Object:            "response",
		CreatedAt:         time.Now().Unix(),
		Model:             model,
		Status:            "completed",
		Error:             nil,
		IncompleteDetails: nil,
		Instructions:      nil,
		MaxOutputTokens:   nil,
		Metadata:          map[string]interface{}{},
		ParallelToolCalls: true,
		PreviousResponseID: nil,
		Temperature:       1,
		ToolChoice:        "auto",
		Tools:             []interface{}{},
		TopP:              1,
		Truncation:        "disabled",
		Output: []types.ResponseOutput{
			{
				Type:   "message",
				ID:     "msg-" + uuid.New().String(),
				Status: "completed",
				Role:   "assistant",
				Content: []types.OutputContent{
					{
						Type:        "output_text",
						Text:        content,
						Annotations: []interface{}{},
					},
				},
			},
		},
		Usage: types.ResponsesUsage{
			InputTokens:  0,
			OutputTokens: 0,
			TotalTokens:  0,
		},
	}
}

func extractPrompt(input interface{}) string {
	if input == nil {
		return ""
	}
	switch v := input.(type) {
	case string:
		return v
	case []interface{}:
		var msgs []types.Message
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				role, _ := m["role"].(string)
				content, _ := m["content"].(string)
				msgs = append(msgs, types.Message{Role: role, Content: content})
			}
		}
		return formatter.BuildPrompt(msgs)
	}
	return ""
}
