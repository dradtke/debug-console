package nvim

import (
	"fmt"
	"log"
	"strings"

	"github.com/dradtke/debug-console/dap"
	"github.com/dradtke/debug-console/types"
	"github.com/dradtke/debug-console/util"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

func RegisterCommands(p *plugin.Plugin, d *dap.DAP) {
	p.HandleCommand(&plugin.CommandOptions{
		Name:  "DebugRun",
		NArgs: "*",
		Eval:  "*",
	}, DebugRun(d))
	p.HandleCommand(&plugin.CommandOptions{Name: "ToggleBreakpoint"}, ToggleBreakpoint(d))
	p.HandleCommand(&plugin.CommandOptions{Name: "CurrentLocation"}, CurrentLocation(d))
}

func OnDapExit(v *nvim.Nvim) func() {
	return func() {
		RemoveAllSigns(v, SignGroupCurrentLocation)
	}
}

func DebugRun(d *dap.DAP) any {
	return func(v *nvim.Nvim, args []string, eval *struct {
		Path     string `eval:"expand('%:p')"`
		Filetype string `eval:"getbufvar(bufnr('%'), '&filetype')"`
	}) error {
		if len(args) == 0 {
			Notify(v, fmt.Sprintf("available configurations: %s", strings.Join(d.Configs.Available(eval.Filetype), ", ")), nvim.LogWarnLevel)
			return nil
		}
		log.Print("Starting debug run")
		launchConfigName := args[0]
		launchConfigArgs := args[1:]
		dapConfig, err := d.Configs.Get(eval.Filetype, launchConfigName)
		if err != nil {
			return fmt.Errorf("No DAP configuration found for filetype: %s", eval.Filetype)
		}
		go func() {
			defer util.LogPanic()
			p, err := d.Run(dapConfig, OnDapExit(v))
			if err != nil {
				errmsg := fmt.Sprintf("Error starting debug adapter: %s", err)
				Notify(v, errmsg, nvim.LogErrorLevel)
				log.Print(errmsg)
				return
			}

			log.Print("Getting launch arguments")

			var launchArgs map[string]any
			if err := v.ExecLua(dapConfig.Launch, &launchArgs, eval.Path, launchConfigArgs); err != nil {
				log.Printf("Error getting launch arguments: %s", err)
				return
			}

			log.Print("Sending launch request")

			if _, err = p.SendRequest(types.NewLaunchRequest(launchArgs)); err != nil {
				log.Printf("Error executing launch request: %s", err)
				return
			}

			log.Print("Sending the configuration")

			if err = SendConfiguration(v, p); err != nil {
				log.Printf("Error sending configuration: %s", err)
				return
			}
		}()
		return nil
	}
}

func ToggleBreakpoint(d *dap.DAP) any {
	return func(v *nvim.Nvim) error {
		lineNum, err := GetLineNumber(v)
		if err != nil {
			return fmt.Errorf("ToggleBreakpoint: %w", err)
		}

		sign, err := GetSignAt(v, SignGroupBreakpoint, "%", lineNum)
		if err != nil {
			return fmt.Errorf("ToggleBreakpoint: %w", err)
		}

		if sign.Exists {
			if err := RemoveSign(v, sign); err != nil {
				return fmt.Errorf("ToggleBreakpoint: %w", err)
			}
		} else {
			if err := PlaceSign(v, SignNameBreakpoint, sign, 98); err != nil {
				return fmt.Errorf("ToggleBreakpoint: %w", err)
			}
		}

		return nil
	}
}

func CurrentLocation(d *dap.DAP) any {
	return func(v *nvim.Nvim) error {
		d.Lock()
		defer d.Unlock()
		if d.StoppedLocation == nil || d.StoppedLocation.Source == nil || d.StoppedLocation.Source.Path == nil {
			Notify(v, "No stopped location", nvim.LogWarnLevel)
			return nil
		}
		return v.Command(fmt.Sprintf("keepalt edit +%d %s", d.StoppedLocation.Line, *d.StoppedLocation.Source.Path))
	}
}
