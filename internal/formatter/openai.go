package formatter

import (
	"time"

	"github.com/google/uuid"

	"github.com/dutakey/shellgate/internal/types"
)

func NewCompletionID() string {
	return "chatcmpl-" + uuid.New().String()
}

func BuildResponse(id, model, content string) *types.ChatCompletionResponse {
	return &types.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []types.Choice{
			{
				Index: 0,
				Message: types.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: types.Usage{
			PromptTokens:     -1,
			CompletionTokens: -1,
			TotalTokens:      -1,
		},
	}
}

func BuildChunk(id, model, content string, finishReason *string) *types.ChatCompletionChunk {
	return &types.ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []types.ChunkChoice{
			{
				Index:        0,
				Delta:        types.Delta{Content: content},
				FinishReason: finishReason,
			},
		},
	}
}

func BuildFirstChunk(id, model string) *types.ChatCompletionChunk {
	return &types.ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []types.ChunkChoice{
			{
				Index:        0,
				Delta:        types.Delta{Role: "assistant"},
				FinishReason: nil,
			},
		},
	}
}

func BuildStopChunk(id, model string) *types.ChatCompletionChunk {
	stop := "stop"
	return &types.ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []types.ChunkChoice{
			{
				Index:        0,
				Delta:        types.Delta{},
				FinishReason: &stop,
			},
		},
	}
}

// BuildPrompt concatenates OpenAI messages into a single prompt string for codex exec.
func BuildPrompt(messages []types.Message) string {
	var parts []string
	for _, m := range messages {
		if m.Content == "" {
			continue
		}
		switch m.Role {
		case "system":
			parts = append(parts, "[system]: "+m.Content)
		case "user":
			parts = append(parts, "[user]: "+m.Content)
		case "assistant":
			parts = append(parts, "[assistant]: "+m.Content)
		default:
			parts = append(parts, m.Content)
		}
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "\n\n"
		}
		result += p
	}
	return result
}

