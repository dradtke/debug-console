package dap

import (
	"encoding/json"
	"errors"
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

func (d *DAP) Next(granularity string) error {
	d.Lock()
	defer d.Unlock()
	if d.StoppedThreadID == 0 {
		return errors.New("no stopped thread!")
	}
	_, err := d.Conn.SendRequest(types.NewNextRequest(types.NextArguments{
		ThreadID:            d.StoppedThreadID,
		SteppingGranularity: granularity,
	}))
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

func (p *Conn) Threads() ([]types.Thread, error) {
	resp, err := p.SendRequest(types.NewThreadsRequest())
	if err != nil {
		return nil, err
	}

	var body types.ThreadsResponse
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("Error parsing threads response: %w", err)
	}

	return body.Threads, nil
}

func (p *Conn) Completions(args types.CompletionsArguments) ([]types.CompletionItem, error) {
	resp, err := p.SendRequest(types.NewCompletionsRequest(args))
	if err != nil {
		return nil, err
	}

	var body types.CompletionsResponse
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("Error parsing completions response: %w", err)
	}
	return body.Targets, nil
}

func (d *DAP) Terminate() error {
	if !d.Capabilities.SupportTerminateDebuggee {
		return types.ErrUnsupported
	}
	d.Lock()
	_, err := d.Conn.SendRequest(types.NewTerminateRequest(types.TerminateArguments{}))
	d.Unlock()
	if err == nil {
		d.Stop()
	}
	return err
}

func (d *DAP) Disconnect() error {
	d.Lock()
	_, err := d.Conn.SendRequest(types.NewDisconnectRequest(types.DisconnectArguments{}))
	d.Unlock()
	if err == nil {
		d.Stop()
	}
	return err
}

// TODO: Add more request types here
