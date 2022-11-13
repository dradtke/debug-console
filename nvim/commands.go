package nvim

import (
	"fmt"
	"log"

	"github.com/dradtke/debug-console/dap"
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
	//p.HandleCommand(&plugin.CommandOptions{Name: "DebugConsoleTest"}, Test)
}

//func Test(v *nvim.Nvim) error {
//	log.Printf("escaped shell value: %s", ShellEscape(v)("hello world"))
//	return nil
//}

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
		defer util.Recover()
		if len(args) == 0 {
			// Notify(v, fmt.Sprintf("available configurations for filetype '%s': %s", eval.Filetype, strings.Join(d.Configs.Available(eval.Filetype), ", ")), nvim.LogWarnLevel)
			Notify(v, "No debug configuration specified", nvim.LogWarnLevel)
			return nil
		}
		log.Print("Starting debug run")
		luaRequire := fmt.Sprintf("require('debug-console.%s.%s')", eval.Filetype, args[0])
		d.Lock()
		d.LaunchArgs.Filepath = eval.Path
		d.LaunchArgs.UserArgs = args[1:]
		d.LaunchArgs.LaunchFunc = luaRequire + ".launch"
		d.Unlock()
		return v.ExecLua(luaRequire + ".run()", nil)
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
