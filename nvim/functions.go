package nvim

import (
	"errors"
	"fmt"
	"log"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"

	"github.com/dradtke/debug-console/dap"
	"github.com/dradtke/debug-console/types"
)

func RegisterFunctions(p *plugin.Plugin, d *dap.DAP) {
	// TODO: define a function that can be used to cancel a run + launch
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleRun"}, Run(d))
	p.HandleFunction(&plugin.FunctionOptions{Name: "DebugConsoleLaunch"}, Launch(d))
}

func Run(d *dap.DAP) any {
	return func(v *nvim.Nvim, args []dap.RunArgs) error {
		if len(args) != 1 {
			return errors.New("expected exactly one argument")
		}
		if _, err := d.Run(args[0], OnDapExit(v)); err != nil {
			return err
		}
		d.RLock()
		launchArgs := d.LaunchArgs
		d.RUnlock()
		return v.ExecLua(launchArgs.LaunchFunc + "(...)", nil, launchArgs.Filepath, launchArgs.UserArgs)
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
				errmsg := fmt.Sprintf("Error executing launch request: %s", err)
				log.Println(errmsg)
				Notify(v, errmsg, nvim.LogErrorLevel)
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
