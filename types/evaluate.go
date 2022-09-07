package types

type EvaluateRequest struct {
	Expression string `json:"expression"`
	Context    string `json:"context"`
}

type EvaluateResponse struct {
	Result string `json:"result"`
}
