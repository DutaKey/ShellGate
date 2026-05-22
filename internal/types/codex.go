package types

import "encoding/json"

// CodexEvent — JSONL event emitted by `codex exec --json`
type CodexEvent struct {
	Type string          `json:"type"`
	Raw  json.RawMessage `json:"-"`
}

// CodexItemEvent — for item.started / item.completed
type CodexItemEvent struct {
	Type string     `json:"type"`
	Item CodexItem  `json:"item"`
}

type CodexItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// CodexErrorEvent — for turn.failed / error
type CodexErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e *CodexErrorEvent) ErrorMessage() string {
	if e.Error != nil && e.Error.Message != "" {
		return e.Error.Message
	}
	return e.Message
}
