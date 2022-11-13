package types

type ReverseRequest struct {
	Seq       int64          `json:"seq"`
	Type      string         `json:"type"`
	Command   string         `json:"command"`
	Arguments map[string]any `json:"arguments"`
}
