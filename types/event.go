package types

import "encoding/json"

type Event struct {
	Seq   int64           `json:"seq"`
	Type  string          `json:"type"`
	Event string          `json:"event"`
	Body  json.RawMessage `json:"body"`
}

type EventHandler func(Event)

// https://microsoft.github.io/debug-adapter-protocol/specification#Events_Stopped
type StoppedEvent struct {
	AllThreadsStopped *bool   `json:"allThreadsStopped"`
	Reason            string `json:"reason"`
	Description       *string `json:"description"`
	ThreadID          *int    `json:"threadId"`
	PreserveFocusHint *bool   `json:"preserveFocusHint"`
	Text              *string `json:"text"`
	HitBreakpointIds  []int  `json:"hitBreakpointIds"`
}

type OutputEvent struct {
	Category string `json:"category"`
	Output   string `json:"output"`
}
