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
		Run(eval.Path, dap.GoStart)
	default:
		return fmt.Errorf("unsupported filetype: %s", eval.Filetype)
	}
	return nil
}

func ToggleBreakpoint(v *nvim.Nvim) error {
	const (
		signGroup = "debug-console-breakpoint"
		signName = "debug-console-breakpoint"
	)
	var lineNum int
	if err := v.Call("line", &lineNum, "."); err != nil {
		return fmt.Errorf("failed to call line(): %w", err)
	}

	var placedSigns []map[string]any
	if err := v.Call("sign_getplaced", &placedSigns, "%", map[string]any{
		"group": "debug-console-breakpoint",
		"lnum": lineNum,
	}); err != nil {
		return fmt.Errorf("failed to call sign_getplaced(): %w", err)
	}

	placedSignDetails := placedSigns[0]["signs"].([]any) 
	if len(placedSignDetails) == 0 {
		log.Println("placing new sign")
		return v.Call("sign_place", nil, 0, signGroup, signName, "%", map[string]any{
			"lnum":     lineNum,
			"priority": 99,
		})
	} else {
		log.Print("removing sign")
		existing := placedSignDetails[0].(map[string]any)
		return v.Call("sign_unplace", nil, signGroup, map[string]any{
			"buffer": placedSigns[0]["bufnr"],
			"id": existing["id"],
		})
	}
}
