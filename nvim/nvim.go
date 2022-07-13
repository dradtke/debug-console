package nvim

import (
	"fmt"
	"log"
	"os"

	"github.com/dradtke/debug-console/dap"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

var (
	dapDir string
)

type StartFunc func(string, dap.Handlers) (*dap.Process, dap.OnInitializedFunc, error)

func Run(filepath string, start StartFunc) {
	if state.Running {
		log.Print("A debug adapter is already running")
		return
	}
	log.Printf("Starting debug adapter...")
	p, onInitialized, err := start(dapDir, dap.Handlers{
		Response: HandleResponse,
		Event:    HandleEvent,
	})
	if err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}
	state = State{
		Running:       true,
		Process:       p,
		Capabilities:  nil,
		OnInitialized: onInitialized,
		Filepath:      filepath,
	}
	go func() {
		if err := state.Process.Wait(); err != nil {
			log.Printf("Debug adapter exited with error: %s", err)
		} else {
			log.Print("Debug adapter exited")
		}
		state = State{}
	}()
	log.Printf("Started debug adapter")

	state.Process.Initialize()
}

func DebugRun(v *nvim.Nvim, eval *struct {
	Path     string `eval:"expand('%:p')"`
	Filetype string `eval:"getbufvar(bufnr('%'), '&filetype')"`
}) error {
	log.Print("starting debug run")
	switch eval.Filetype {
	case "go":
		Run(eval.Path, dap.GoStart)
	default:
		return fmt.Errorf("unsupported filetype: %s", eval.Filetype)
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
	if os.Getenv("NVIM") != "" {
		if err := setLogOutput(); err != nil {
			log.Print(err)
		}

		dapDir = os.Getenv("DAP_DIR")
		if _, err := os.Stat(dapDir); os.IsNotExist(err) {
			if err = os.MkdirAll(dapDir, 0644); err != nil {
				return fmt.Errorf("failed to create dap cache dir: %w", err)
			}
		}
	}

	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleCommand(&plugin.CommandOptions{
			Name: "DebugRun",
			Eval: "*",
		}, DebugRun)
		return nil
	})

	return nil
}
