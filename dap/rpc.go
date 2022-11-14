package dap

import "github.com/dradtke/debug-console/types"

type DAPService struct {
	d *DAP
}

func (r DAPService) Continue(_ struct{}, _ *struct{}) error {
	return r.d.Continue()
}

func (r DAPService) StepIn(_ struct{}, _ *struct{}) error {
	return r.d.StepIn()
}

func (r DAPService) StepOut(_ struct{}, _ *struct{}) error {
	return r.d.StepOut()
}

func (r DAPService) StepBack(_ struct{}, _ *struct{}) error {
	return r.d.StepBack()
}

func (r DAPService) Next(granularity string, _ *struct{}) error {
	return r.d.Next(granularity)
}

func (r DAPService) Evaluate(args types.EvaluateArguments, result *string) error {
	if r.d.StoppedLocation != nil {
		args.FrameID = r.d.StoppedLocation.ID
	}
	v, err := r.d.Conn.Evaluate(args)
	if err != nil {
		return err
	}
	*result = v
	return nil
}

func (r DAPService) Threads(_ struct{}, result *[]types.Thread) error {
	v, err := r.d.Conn.Threads()
	if err != nil {
		return err
	}
	*result = v
	return nil
}

func (r DAPService) Terminate(_ struct{}, _ *struct{}) error {
	return r.d.Terminate()
}

func (r DAPService) Disconnect(_ struct{}, _ *struct{}) error {
	return r.d.Disconnect()
}

func (r DAPService) Capabilities(_ struct{}, capabilities *types.Capabilities) error {
	if r.d.Capabilities != nil {
		*capabilities = *r.d.Capabilities
	}
	return nil
}

func (r DAPService) Completions(args types.CompletionsArguments, results *[]types.CompletionItem) error {
	items, err := r.d.Conn.Completions(args)
	if err != nil {
		return err
	}
	*results = items
	return nil
}
