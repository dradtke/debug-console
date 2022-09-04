package dap

type Stopped struct {
	AllThreadsStopped bool   `json:"allThreadsStopped"`
	Reason            string `json:"reason"`
	ThreadID          int    `json:"threadId"`
}
