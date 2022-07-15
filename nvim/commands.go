package nvim

import (
	"fmt"
	"log"

	"github.com/dradtke/debug-console/dap"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

func RegisterCommands(p *plugin.Plugin) {
	p.HandleCommand(&plugin.CommandOptions{Name: "DebugRun", Eval: "*"}, DebugRun)
	p.HandleCommand(&plugin.CommandOptions{Name: "ToggleBreakpoint"}, ToggleBreakpoint)
}

func DebugRun(v *nvim.Nvim, eval *struct {
	Path     string `eval:"expand('%:p')"`
	Filetype string `eval:"getbufvar(bufnr('%'), '&filetype')"`
}) error {
	log.Print("Starting debug run")
	switch eval.Filetype {
	case "go":
		go func() {
			p, err := dap.Run(dap.RunArgs{
				Filepath:     eval.Path,
				DapDir:       dapDir,
				EventHandler: HandleEvent,
				DapCommand:   dap.GoCommand,
			})
			if err != nil {
				log.Printf("Error starting debug adapter: %s", err)
				return
			}
			log.Print("Go debug adapter initialized, launching")
			if _, err := dap.GoLaunch(eval.Path, p); err != nil {
				log.Printf("Error launching Go: %s", err)
				return
			}
			if err := SendConfiguration(v, p); err != nil {
				log.Printf("Error sending configuration: %s", err)
			}
		}()
	default:
		return fmt.Errorf("unsupported filetype: %s", eval.Filetype)
	}
	return nil
}

func ToggleBreakpoint(v *nvim.Nvim) error {
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
		// ???: What name should be used here?
		if err := PlaceSign(v, SignNameBreakpoint, sign); err != nil {
			return fmt.Errorf("ToggleBreakpoint: %w", err)
		}
	}

	return nil
}
