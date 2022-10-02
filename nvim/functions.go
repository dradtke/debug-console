package nvim

import (
	"errors"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"

	"github.com/dradtke/debug-console/dap"
)

func RegisterFunctions(p *plugin.Plugin, d *dap.DAP) {
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleSetDefaultConfig"}, SetDefaultConfig(d))
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleSetUserConfig"}, SetUserConfig(d))
}

func SetDefaultConfig(d *dap.DAP) any {
	return func(v *nvim.Nvim, args []map[string]dap.ConfigMap) error {
		if len(args) != 1 {
			return errors.New("expected exactly one argument")
		}
		d.Lock()
		defer d.Unlock()
		d.Configs.Defaults = args[0]
		return nil
	}
}

func SetUserConfig(d *dap.DAP) any {
	return func(v *nvim.Nvim, args []dap.ConfigMap) error {
		if len(args) != 1 {
			return errors.New("expected exactly one argument")
		}
		d.Lock()
		defer d.Unlock()
		d.Configs.User = args[0]
		return nil
	}
}
