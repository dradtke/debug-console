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
