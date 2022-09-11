package nvim

import (
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"

	"github.com/dradtke/debug-console/dap"
)

func RegisterFunctions(p *plugin.Plugin, d *dap.DAP) {
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleSetConfig"}, SetConfig(d))
}

func SetConfig(d *dap.DAP) any {
	return func(v *nvim.Nvim) error {
		Notify(v, "Setting debug console config", nvim.LogInfoLevel)
		return nil
	}
}
