package nvim

import (
	"fmt"
	"log"
	"os"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

var logOutputSet = false

func cmdDebugRun(v *nvim.Nvim) error {
	log.Print("Logging from DebugRun")
	if err := v.Notify("Calling Notify from DebugRun", nvim.LogInfoLevel, make(map[string]interface{})); err != nil {
		return err
	}
	return nil
}

func setLogOutput() error {
	filename := os.Getenv("LOG_FILE")
	if filename == "" {
		return nil
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("setLogOutput: error opening log file: %w", err)
	}
	log.SetFlags(0)
	log.SetOutput(f)
	return nil
}

func Main() error {
	if err := setLogOutput(); err != nil {
		log.Print(err)
	}

	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleCommand(&plugin.CommandOptions{Name: "DebugRun"}, cmdDebugRun)
		return nil
	})
	return nil
}
