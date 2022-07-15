package dap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

var (
	// Current state global
	state   State
	stateMu sync.Mutex
)

type State struct {
	Running      bool
	Process      *Process
	Capabilities map[string]bool
	Filepath     string
}

type RunArgs struct {
	DapDir, Filepath string
	EventHandler     func(Event)
	DapCommand       DapCommandFunc
}

type DapCommandFunc func(string) ([]string, error)

func Run(args RunArgs) (*Process, error) {
	if state.Running {
		return nil, errors.New("A debug adapter is already running")
	}

	log.Printf("Starting debug adapter...")
	command, err := args.DapCommand(args.DapDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to build DAP command: %w", err)
	}

	p, err := NewProcess(args.EventHandler, command[0], command[1:]...)
	if err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}

	stateMu.Lock()
	state = State{
		Running:      true,
		Process:      p,
		Capabilities: nil,
		Filepath:     args.Filepath,
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

	resp, err := p.Initialize()
	if err != nil {
		return p, fmt.Errorf("Error initializing debug adapter: %w", err)
	}
	state.Capabilities = make(map[string]bool)
	if err := json.Unmarshal(resp.Body, &state.Capabilities); err != nil {
		log.Printf("Error parsing capabilities: %s", err)
	}

	return state.Process, nil
}

func ClearState() {
	state = State{}
}
