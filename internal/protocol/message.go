package protocol

import "encoding/json"

const Version = "1.0"

type CoreInfo struct {
	Version string `json:"version"`
}

type ModInfo struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type ErrorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type CallTarget struct {
	Capability string `json:"capability,omitempty"`
	Mod        string `json:"mod,omitempty"`
	Action     string `json:"action,omitempty"`
}

type Message struct {
	Type      string          `json:"type"`
	Protocols []string        `json:"protocols,omitempty"`
	Protocol  string          `json:"protocol,omitempty"`
	Core      *CoreInfo       `json:"core,omitempty"`
	Mod       *ModInfo        `json:"mod,omitempty"`
	ID        string          `json:"id,omitempty"`
	Parent    string          `json:"parent,omitempty"`
	Action    string          `json:"action,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Event     string          `json:"event,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Output    json.RawMessage `json:"output,omitempty"`
	Error     *ErrorPayload   `json:"error,omitempty"`
	Target    *CallTarget     `json:"target,omitempty"`
}

type Event struct {
	InvocationID string
	Name         string
	Data         json.RawMessage
}

type CallRequest struct {
	ID     string
	Parent string
	Target CallTarget
	Input  json.RawMessage
}
