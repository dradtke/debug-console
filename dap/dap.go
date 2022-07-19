package dap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/dradtke/debug-console/tmux"
)

type DAP struct {
	sync.RWMutex

	// Dir is where debug adapters are saved locally.
	Dir          string
	EventHandler func(Event)
	Conn         *Conn
	Capabilities map[string]bool
}

type DapCommandFunc func(string) ([]string, error)

func (d *DAP) Run(command []string) (c *Conn, err error) {
	d.RLock()
	if d.Conn != nil {
		defer d.RUnlock()
		return nil, errors.New("A debug adapter is already running")
	}
	d.RUnlock()

	log.Printf("Starting debug adapter...")

	if c, err = d.NewProcess(command[0], command[1:]...); err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}

	defer func() {
		if err != nil && c != nil {
			c.Stop()
		}
	}()

	d.Lock()
	d.Conn = c
	d.Unlock()

	go func() {
		if err := c.Wait(); err != nil {
			// ???: Suppress this message if the adapter was killed by Neovim exiting?
			log.Printf("Debug adapter exited with error: %s", err)
		} else {
			log.Print("Debug adapter exited")
		}
		d.ClearProcess()
	}()

	log.Printf("Started debug adapter")
	numPanes, err := tmux.NumPanes()
	if err != nil {
		return nil, fmt.Errorf("Error getting number of tmux panes: %w", err)
	}

	if numPanes == 1 {
		if err = tmux.Split(); err != nil {
			return nil, fmt.Errorf("Error splitting tmux panes: %w", err)
		}
	}

	resp, err := c.Initialize()
	if err != nil {
		return c, fmt.Errorf("Error initializing debug adapter: %w", err)
	}
	d.Capabilities = make(map[string]bool)
	if err := json.Unmarshal(resp.Body, &d.Capabilities); err != nil {
		log.Printf("Error parsing capabilities: %s", err)
	}

	return c, nil
}

func (d *DAP) SendRequest(name string, args any) (Response, error) {
	d.Lock()
	p := d.Conn
	d.Unlock()
	if p == nil {
		return Response{}, errors.New("No process running")
	}
	return p.SendRequest(name, args)
}

func (d *DAP) ClearProcess() {
	d.Lock()
	d.Conn = nil
	d.Unlock()
}
