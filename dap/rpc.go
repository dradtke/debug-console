package dap

type DAPService struct {
	d *DAP
}

func (r DAPService) Continue(_ struct{}, _ *struct{}) error {
	return r.d.Continue()
}
