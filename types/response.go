package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Response struct {
	Seq        int64           `json:"seq,omitempty"`
	Type       string          `json:"type"`
	RequestSeq int64           `json:"request_seq"`
	Command    string          `json:"command"`
	Success    bool            `json:"success"`
	Body       json.RawMessage `json:"body,omitempty"`
}

type ErrorResponse struct {
	Details struct {
		ID        int               `json:"id"`
		Format    string            `json:"format"`
		Variables map[string]string `json:"variables"`
	} `json:"error"`
	ShowUser bool `json:"showUser"`
}

func (r ErrorResponse) Error() string {
	msg := r.Details.Format
	for name, value := range r.Details.Variables {
		msg = strings.ReplaceAll(msg, "{"+name+"}", value)
	}

	return fmt.Sprintf("%d: %s", r.Details.ID, msg)
}

type EvaluateResponse struct {
	Result string `json:"result"`
}

type ThreadsResponse struct {
	Threads []Thread `json:"threads"`
}

type CompletionsResponse struct {
	Targets []CompletionItem `json:"targets"`
}

type RunInTerminalResponse struct {
	ProcessID      int `json:"processId,omitempty"`
	ShellProcessID int `json:"shellProcessId,omitempty"`
}

func NewResponse(requestSeq int64, command string, success bool, body any) Response {
	b, err := json.Marshal(body)
	if err != nil {
		panic("types.NewResponse: failed to marshal body: " + err.Error())
	}
	return Response{
		Type:       "response",
		RequestSeq: requestSeq,
		Command:    command,
		Success:    success,
		Body:       json.RawMessage(b),
	}
}
