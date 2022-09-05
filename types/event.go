package types

import "encoding/json"

type Event struct {
	Seq   int64           `json:"seq"`
	Type  string          `json:"type"`
	Event string          `json:"event"`
	Body  json.RawMessage `json:"body"`
}

type EventHandler func(Event)
