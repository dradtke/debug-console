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
	SupportsConfigurationDoneRequest      bool                         `json:"supportsConfigurationDoneRequest,omitempty"`
	SupportsFunctionBreakpoints           bool                         `json:"supportsFunctionBreakpoints,omitempty"`
	SupportsConditionalBreakpoints        bool                         `json:"supportsConditionalBreakpoints,omitempty"`
	SupportsHitConditionalBreakpoints     bool                         `json:"supportsHitConditionalBreakpoints,omitempty"`
	SupportsEvaluateForHovers             bool                         `json:"supportsEvaluateForHovers,omitempty"`
	ExceptionBreakpointFilters            []ExceptionBreakpointsFilter `json:"exceptionBreakpointFilters,omitempty"`
	SupportsStepBack                      bool                         `json:"supportsStepBack,omitempty"`
	SupportsSetVariable                   bool                         `json:"supportsSetVariable,omitempty"`
	SupportsRestartFrame                  bool                         `json:"supportsRestartFrame,omitempty"`
	SupportsGotoTargetsRequest            bool                         `json:"supportsGotoTargetsRequest,omitempty"`
	SupportsStepInTargetsRequest          bool                         `json:"supportsStepInTargetsRequest,omitempty"`
	SupportsCompletionsRequest            bool                         `json:"supportsCompletionsRequest,omitempty"`
	CompletionTriggerCharacters           []string                     `json:"completionTriggerCharacters,omitempty"`
	SupportsModulesRequest                bool                         `json:"supportsModulesRequest,omitempty"`
	AdditionalModuleColumns               []ColumnDescriptor           `json:"additionalModuleColumns,omitempty"`
	SupportedChecksumAlgorithms           []ChecksumAlgorithm          `json:"supportedChecksumAlgorithms,omitempty"`
	SupportsRestartRequest                bool                         `json:"supportsRestartRequest,omitempty"`
	SupportsExceptionOptions              bool                         `json:"supportsExceptionOptions,omitempty"`
	SupportsValueFormattingOptions        bool                         `json:"supportsValueFormattingOptions,omitempty"`
	SupportsExceptionInfoRequest          bool                         `json:"supportsExceptionInfoRequest,omitempty"`
	SupportTerminateDebuggee              bool                         `json:"supportTerminateDebuggee,omitempty"`
	SupportSuspendDebuggee                bool                         `json:"supportSuspendDebuggee,omitempty"`
	SupportsDelayedStackTraceLoading      bool                         `json:"supportsDelayedStackTraceLoading,omitempty"`
	SupportsLoadedSourcesRequest          bool                         `json:"supportsLoadedSourcesRequest,omitempty"`
	SupportsLogPoints                     bool                         `json:"supportsLogPoints,omitempty"`
	SupportsTerminateThreadsRequest       bool                         `json:"supportsTerminateThreadsRequest,omitempty"`
	SupportsSetExpression                 bool                         `json:"supportsSetExpression,omitempty"`
	SupportsTerminateRequest              bool                         `json:"supportsTerminateRequest,omitempty"`
	SupportsDataBreakpoints               bool                         `json:"supportsDataBreakpoints,omitempty"`
	SupportsReadMemoryRequest             bool                         `json:"supportsReadMemoryRequest,omitempty"`
	SupportsWriteMemoryRequest            bool                         `json:"supportsWriteMemoryRequest,omitempty"`
	SupportsDisassembleRequest            bool                         `json:"supportsDisassembleRequest,omitempty"`
	SupportsCancelRequest                 bool                         `json:"supportsCancelRequest,omitempty"`
	SupportsBreakpointLocationsRequest    bool                         `json:"supportsBreakpointLocationsRequest,omitempty"`
	SupportsClipboardContext              bool                         `json:"supportsClipboardContext,omitempty"`
	SupportsSteppingGranularity           bool                         `json:"supportsSteppingGranularity,omitempty"`
	SupportsInstructionBreakpoints        bool                         `json:"supportsInstructionBreakpoints,omitempty"`
	SupportsExceptionFilterOptions        bool                         `json:"supportsExceptionFilterOptions,omitempty"`
	SupportsSingleThreadExecutionRequests bool                         `json:"supportsSingleThreadExecutionRequests,omitempty"`
}

type ExceptionBreakpointsFilter struct{}
type ColumnDescriptor struct{}
type ChecksumAlgorithm struct{}

type Thread struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CompletionItem struct {
	Label           string `json:"label"`
	Type            string `json:"type"`
	Text            string `json:"text,omitempty"`
	SortText        string `json:"sortText,omitempty"`
	Detail          string `json:"detail,omitempty"`
	Start           *int   `json:"start,omitempty"`
	Length          int    `json:"length,omitempty"`
	SelectionStart  *int   `json:"selectionStart,omitempty"`
	SelectionLength int    `json:"selectionLength,omitempty"`
}
