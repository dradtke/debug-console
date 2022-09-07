package dap

import (
	"encoding/json"
	"fmt"

	"github.com/dradtke/debug-console/types"
)

func (p *Conn) Initialize() (types.Response, error) {
	return p.SendRequest(types.NewInitializeRequest(types.InitializeArguments{
		AdapterID:                    "debug-console",
		PathFormat:                   "path",
		LinesStartAt1:                true,
		ColumnsStartAt1:              true,
		SupportsRunInTerminalRequest: true,
	}))
}

func (p *Conn) ConfigurationDone() (types.Response, error) {
	return p.SendRequest(types.NewConfigurationDoneRequest())
}

func (d *DAP) Continue() error {
	d.Lock()
	defer d.Unlock()
	_, err := d.Conn.SendRequest(types.NewContinueRequest())
	if err == nil {
		d.StoppedLocation = nil
	}
	return err
}

func (p *Conn) Evaluate(args types.EvaluateArguments) (string, error) {
	resp, err := p.SendRequest(types.NewEvaluateRequest(args))
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
