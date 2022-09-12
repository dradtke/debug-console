package nvim

import (
	"errors"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"

	"github.com/dradtke/debug-console/dap"
)

func RegisterFunctions(p *plugin.Plugin, d *dap.DAP) {
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleSetConfig"}, SetConfig(d))
}

func SetConfig(d *dap.DAP) any {
	return func(v *nvim.Nvim, args []dap.ConfigMap) error {
		if len(args) != 1 {
			return errors.New("expected exactly one argument")
		}
		d.Lock()
		defer d.Unlock()
		d.ConfigMap = args[0]
		return nil
	}
}
