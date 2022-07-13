package nvim

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dradtke/debug-console/dap"
	"github.com/neovim/go-client/nvim/plugin"
)

const (
	SignGroupBreakpoint = "debug-console-breakpoint"
	SignNameBreakpoint = "debug-console-breakpoint"
)

var (
	dapDir string
)

type StartFunc func(string, func(dap.Event)) (*dap.Process, error)

func Run(filepath string, start StartFunc, onInitialized func(*dap.Process)) {
	if state.Running {
		log.Print("A debug adapter is already running")
		return
	}
	log.Printf("Starting debug adapter...")
	p, err := start(dapDir, HandleEvent)
	if err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}
	stateMu.Lock()
	state = State{
		Running:       true,
		Process:       p,
		Capabilities:  nil,
		Filepath:      filepath,
	}
	stateMu.Unlock()
	go func() {
		if err := p.Wait(); err != nil {
			log.Printf("Debug adapter exited with error: %s", err)
		} else {
			log.Print("Debug adapter exited")
		}
		stateMu.Lock()
		state = State{}
		stateMu.Unlock()
	}()
	log.Printf("Started debug adapter")

	go func() {
		resp, err := p.Initialize()
		if err != nil {
			log.Printf("Error initializing debug adapter: %s", err)
			return
		}
		state.Capabilities = make(map[string]bool)
		if err := json.Unmarshal(resp.Body, &state.Capabilities); err != nil {
			log.Printf("Error parsing capabilities: %s", err)
		}
		onInitialized(state.Process)
	}()
}

func SendConfiguration(p *dap.Process) {
	if _, err := p.SendRequest("configurationDone", make(map[string]any)); err != nil {
		log.Printf("Error finishing configuration: %s", err)
	}
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
		RegisterCommands(p)
		return nil
	})

	return nil
}
