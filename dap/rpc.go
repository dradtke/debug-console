package dap

import "github.com/dradtke/debug-console/types"

type DAPService struct {
	d *DAP
}

func (r DAPService) Continue(_ struct{}, _ *struct{}) error {
	return r.d.Continue()
}

func (r DAPService) Evaluate(args types.EvaluateArguments, resp *types.EvaluateResponse) error {
	result, err := r.d.Conn.Evaluate(args)
	if err != nil {
		return err
	}
	*resp = types.EvaluateResponse{Result: result}
	return nil
}
