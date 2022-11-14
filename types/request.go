package types

type Request interface {
	Seq() int64
	Command() string
}

type request struct {
	SeqField     int64  `json:"seq"`
	Type         string `json:"type"` // ???: always "request"?
	CommandField string `json:"command"`
}

func (r request) Seq() int64 {
	return r.SeqField
}

func (r request) Command() string {
	return r.CommandField
}

func newRequest(command string) request {
	return request{
		SeqField:     NextSeq(),
		Type:         "request",
		CommandField: command,
	}
}

type InitializeArguments struct {
	AdapterID                           string `json:"adapterID"`
	PathFormat                          string `json:"pathFormat"`
	LinesStartAt1                       bool   `json:"linesStartAt1"`
	ColumnsStartAt1                     bool   `json:"columnsStartAt1"`
	SupportsRunInTerminalRequest        bool   `json:"supportsRunInTerminalRequest"`
	SupportsArgsCanBeInterpretedByShell bool   `json:"supportsArgsCanBeInterpretedByShell"`
}

func NewInitializeRequest(args InitializeArguments) Request {
	return struct {
		request
		Arguments InitializeArguments `json:"arguments"`
	}{
		request:   newRequest("initialize"),
		Arguments: args,
	}
}

func NewConfigurationDoneRequest() Request {
	return newRequest("configurationDone")
}

type ContinueArguments struct {
	ThreadID int `json:"threadId,omitempty"`
}

func NewContinueRequest(args ContinueArguments) Request {
	return struct {
		request
		Arguments ContinueArguments `json:"arguments"`
	}{
		request:   newRequest("continue"),
		Arguments: args,
	}
}

func NewStepInRequest() Request {
	return newRequest("stepIn")
}

func NewStepOutRequest() Request {
	return newRequest("stepOut")
}

func NewStepBackRequest() Request {
	return newRequest("stepBack")
}

func NewThreadsRequest() Request {
	return newRequest("threads")
}

type EvaluateArguments struct {
	Expression string `json:"expression"`
	Context    string `json:"context,omitempty"`
	FrameID    int    `json:"frameId,omitempty"`
}

func NewEvaluateRequest(args EvaluateArguments) Request {
	return struct {
		request
		Arguments EvaluateArguments `json:"arguments"`
	}{
		request:   newRequest("evaluate"),
		Arguments: args,
	}
}

type StackTraceArguments struct {
	ThreadID int               `json:"threadId"`
	Levels   int               `json:"levels"`
	Format   *StackFrameFormat `json:"format"`
}

func NewStackTraceRequest(args StackTraceArguments) Request {
	return struct {
		request
		Arguments StackTraceArguments `json:"arguments"`
	}{
		request:   newRequest("stackTrace"),
		Arguments: args,
	}
}

func NewLaunchRequest(args map[string]any) Request {
	return struct {
		request
		Arguments map[string]any `json:"arguments"`
	}{
		request:   newRequest("launch"),
		Arguments: args,
	}
}

type SetBreakpointArguments struct {
	Source      Source             `json:"source"`
	Breakpoints []SourceBreakpoint `json:"breakpoints,omitempty"`
}

func NewSetBreakpointRequest(args SetBreakpointArguments) Request {
	return struct {
		request
		Arguments SetBreakpointArguments `json:"arguments"`
	}{
		request:   newRequest("setBreakpoints"),
		Arguments: args,
	}
}

type TerminateArguments struct {
	Restart bool `json:"restart,omitempty"`
}

func NewTerminateRequest(args TerminateArguments) Request {
	return struct {
		request
		Arguments TerminateArguments `json:"arguments"`
	}{
		request:   newRequest("terminate"),
		Arguments: args,
	}
}

type DisconnectArguments struct {
	Restart           bool `json:"restart,omitempty"`
	TerminateDebuggee bool `json:"terminateDebuggee,omitempty"`
	SuspendDebuggee   bool `json:"suspendDebuggee,omitempty"`
}

func NewDisconnectRequest(args DisconnectArguments) Request {
	return struct {
		request
		Arguments DisconnectArguments `json:"arguments"`
	}{
		request:   newRequest("disconnect"),
		Arguments: args,
	}
}

type CompletionsArguments struct {
	FrameID *int   `json:"frameId,omitempty"`
	Text    string `json:"text"`
	Column  int    `json:"column"`
	Line    *int   `json:"line,omitempty"`
}

func NewCompletionsRequest(args CompletionsArguments) Request {
	return struct {
		request
		Arguments CompletionsArguments `json:"arguments"`
	}{
		request:   newRequest("completions"),
		Arguments: args,
	}
}

type NextArguments struct {
	ThreadID            int    `json:"threadId"`
	SteppingGranularity string `json:"steppingGranularity,omitempty"`
}

func NewNextRequest(args NextArguments) Request {
	return struct {
		request
		Arguments NextArguments `json:"arguments"`
	}{
		request:   newRequest("next"),
		Arguments: args,
	}
}
