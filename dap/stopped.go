package dap

// https://microsoft.github.io/debug-adapter-protocol/specification#Events_Stopped
type Stopped struct {
	AllThreadsStopped *bool   `json:"allThreadsStopped"`
	Reason            string `json:"reason"`
	Description       *string `json:"description"`
	ThreadID          *int    `json:"threadId"`
	PreserveFocusHint *bool   `json:"preserveFocusHint"`
	Text              *string `json:"text"`
	HitBreakpointIds  []int  `json:"hitBreakpointIds"`
}
