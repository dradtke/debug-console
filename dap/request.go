package dap

import "github.com/dradtke/debug-console/types"

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

// TODO: Add more request types here
