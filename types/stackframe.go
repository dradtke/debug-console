package types

type StackFrame struct {
	ID     int     `json:"id"`
	Source *Source `json:"source"`
	Line   int     `json:"line"`
	Column int     `json:"column"`
}
