package dap

import (
	"encoding/json"
	"fmt"

	"github.com/dradtke/debug-console/types"
)

func (p *Conn) Initialize() (types.Response, error) {
	type Args struct {
		AdapterID                    string `json:"adapterID"`
		PathFormat                   string `json:"pathFormat"`
		LinesStartAt1                bool   `json:"linesStartAt1"`
		ColumnsStartAt1              bool   `json:"columnsStartAt1"`
		SupportsRunInTerminalRequest bool   `json:"supportsRunInTerminalRequest"`
	}

	return p.SendRequest("initialize", Args{
		AdapterID:                    "debug-console",
		PathFormat:                   "path",
		LinesStartAt1:                true,
		ColumnsStartAt1:              true,
		SupportsRunInTerminalRequest: true,
	})
}

func (p *Conn) ConfigurationDone() (types.Response, error) {
	return p.SendRequest("configurationDone", struct{}{})
}

func (d *DAP) Continue() error {
	d.Lock()
	defer d.Unlock()
	_, err := d.Conn.SendRequest("continue", nil)
	if err == nil {
		d.StoppedLocation = nil
	}
	return err
}

func (p *Conn) Evaluate(req types.EvaluateRequest) (string, error) {
	resp, err := p.SendRequest("evaluate", req)
	if err != nil {
		return "", err
	}

	var body types.EvaluateResponse
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return "", fmt.Errorf("Error parsing evaluate response: %w", err)
	}

	return body.Result, nil
}

// TODO: Add more request types here
