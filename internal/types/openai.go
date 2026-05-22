package types

// ChatCompletionRequest — OpenAI-compatible request body
type ChatCompletionRequest struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Stream          bool      `json:"stream"`
	// ReasoningEffort: low | medium | high | extra-high
	ReasoningEffort string    `json:"reasoning_effort,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse — non-streaming response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionChunk — streaming SSE chunk
type ChatCompletionChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

type ChunkChoice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ModelsResponse — GET /v1/models
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ResponsesRequest — OpenAI Responses API (POST /v1/responses)
type ResponsesRequest struct {
	Model           string            `json:"model"`
	Input           interface{}       `json:"input"` // string or []Message
	Stream          bool              `json:"stream"`
	ReasoningEffort string            `json:"reasoning_effort,omitempty"`
	Tools           []interface{}     `json:"tools,omitempty"`
}

// ResponsesResponse — OpenAI Responses API response (fully compliant)
type ResponsesResponse struct {
	ID                 string           `json:"id"`
	Object             string           `json:"object"`
	CreatedAt          int64            `json:"created_at"`
	Model              string           `json:"model"`
	Status             string           `json:"status"`
	Output             []ResponseOutput `json:"output"`
	Usage              ResponsesUsage   `json:"usage"`
	Error              interface{}      `json:"error"`
	IncompleteDetails  interface{}      `json:"incomplete_details"`
	Instructions       interface{}      `json:"instructions"`
	MaxOutputTokens    interface{}      `json:"max_output_tokens"`
	Metadata           interface{}      `json:"metadata"`
	ParallelToolCalls  bool             `json:"parallel_tool_calls"`
	PreviousResponseID interface{}      `json:"previous_response_id"`
	Temperature        float64          `json:"temperature"`
	ToolChoice         string           `json:"tool_choice"`
	Tools              []interface{}    `json:"tools"`
	TopP               float64          `json:"top_p"`
	Truncation         string           `json:"truncation"`
}

type ResponseOutput struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Status  string          `json:"status"`
	Role    string          `json:"role"`
	Content []OutputContent `json:"content"`
}

type OutputContent struct {
	Type        string        `json:"type"`
	Text        string        `json:"text"`
	Annotations []interface{} `json:"annotations"`
}

type ResponsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ErrorResponse — OpenAI-compatible error
type ErrorResponse struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}
