package nvim

import (
	"errors"
	"log"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"

	"github.com/dradtke/debug-console/dap"
	"github.com/dradtke/debug-console/types"
)

func RegisterFunctions(p *plugin.Plugin, d *dap.DAP) {
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleSetDefaultConfig"}, SetDefaultConfig(d))
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleSetUserConfig"}, SetUserConfig(d))
	// TODO: define a function that can be used to cancel a run + launch
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleLaunch"}, Launch(d))
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

func Launch(d *dap.DAP) any {
	return func(v *nvim.Nvim, args []map[string]any) error {
		if len(args) != 1 {
			return errors.New("expected exactly one argument")
		}

		d.Lock()
		p := d.Conn
		d.Unlock()

		if p == nil {
			return errors.New("No process found")
		}
		log.Print("Continuing the launch")

		go func() {
			if _, err := p.SendRequest(types.NewLaunchRequest(args[0])); err != nil {
				log.Printf("Error executing launch request: %s", err)
				return
			}

			log.Print("Sending the configuration")

			if err := SendConfiguration(v, p); err != nil {
				log.Printf("Error sending configuration: %s", err)
				return
			}
		}()

		return nil
	}
}
