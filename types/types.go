package types

type Source struct {
	Name *string `json:"name,omitempty"`
	Path *string `json:"path,omitempty"`
}

type SourceBreakpoint struct {
	Line int `json:"line"`
}

type StackFrame struct {
	ID     int     `json:"id"`
	Source *Source `json:"source"`
	Line   int     `json:"line"`
	Column int     `json:"column"`
}

type StackFrameFormat struct {
	Line *bool `json:"line"`
}

type Capabilities struct {
	SupportsConfigurationDoneRequest      bool                         `json:"supportsConfigurationDoneRequest"`
	SupportsFunctionBreakpoints           bool                         `json:"supportsFunctionBreakpoints"`
	SupportsConditionalBreakpoints        bool                         `json:"supportsConditionalBreakpoints"`
	SupportsHitConditionalBreakpoints     bool                         `json:"supportsHitConditionalBreakpoints"`
	SupportsEvaluateForHovers             bool                         `json:"supportsEvaluateForHovers"`
	ExceptionBreakpointFilters            []ExceptionBreakpointsFilter `json:"exceptionBreakpointFilters"`
	SupportsStepBack                      bool                         `json:"supportsStepBack"`
	SupportsSetVariable                   bool                         `json:"supportsSetVariable"`
	SupportsRestartFrame                  bool                         `json:"supportsRestartFrame"`
	SupportsGotoTargetsRequest            bool                         `json:"supportsGotoTargetsRequest"`
	SupportsStepInTargetsRequest          bool                         `json:"supportsStepInTargetsRequest"`
	SupportsCompletionsRequest            bool                         `json:"supportsCompletionsRequest"`
	CompletionTriggerCharacters           []string                     `json:"completionTriggerCharacters"`
	SupportsModulesRequest                bool                         `json:"supportsModulesRequest"`
	AdditionalModuleColumns               []ColumnDescriptor           `json:"additionalModuleColumns"`
	SupportedChecksumAlgorithms           []ChecksumAlgorithm          `json:"supportedChecksumAlgorithms"`
	SupportsRestartRequest                bool                         `json:"supportsRestartRequest"`
	SupportsExceptionOptions              bool                         `json:"supportsExceptionOptions"`
	SupportsValueFormattingOptions        bool                         `json:"supportsValueFormattingOptions"`
	SupportsExceptionInfoRequest          bool                         `json:"supportsExceptionInfoRequest"`
	SupportTerminateDebuggee              bool                         `json:"supportTerminateDebuggee"`
	SupportSuspendDebuggee                bool                         `json:"supportSuspendDebuggee"`
	SupportsDelayedStackTraceLoading      bool                         `json:"supportsDelayedStackTraceLoading"`
	SupportsLoadedSourcesRequest          bool                         `json:"supportsLoadedSourcesRequest"`
	SupportsLogPoints                     bool                         `json:"supportsLogPoints"`
	SupportsTerminateThreadsRequest       bool                         `json:"supportsTerminateThreadsRequest"`
	SupportsSetExpression                 bool                         `json:"supportsSetExpression"`
	SupportsTerminateRequest              bool                         `json:"supportsTerminateRequest"`
	SupportsDataBreakpoints               bool                         `json:"supportsDataBreakpoints"`
	SupportsReadMemoryRequest             bool                         `json:"supportsReadMemoryRequest"`
	SupportsWriteMemoryRequest            bool                         `json:"supportsWriteMemoryRequest"`
	SupportsDisassembleRequest            bool                         `json:"supportsDisassembleRequest"`
	SupportsCancelRequest                 bool                         `json:"supportsCancelRequest"`
	SupportsBreakpointLocationsRequest    bool                         `json:"supportsBreakpointLocationsRequest"`
	SupportsClipboardContext              bool                         `json:"supportsClipboardContext"`
	SupportsSteppingGranularity           bool                         `json:"supportsSteppingGranularity"`
	SupportsInstructionBreakpoints        bool                         `json:"supportsInstructionBreakpoints"`
	SupportsExceptionFilterOptions        bool                         `json:"supportsExceptionFilterOptions"`
	SupportsSingleThreadExecutionRequests bool                         `json:"supportsSingleThreadExecutionRequests"`
}

type ExceptionBreakpointsFilter struct{}
type ColumnDescriptor struct{}
type ChecksumAlgorithm struct{}

type Thread struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
