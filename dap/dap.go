package dap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
)

type DAP struct {
	sync.RWMutex

	// Dir is where debug adapters are saved locally.
	Dir string
	EventHandler func(Event)
	Process      *Process
	Capabilities map[string]bool
}

type DapCommandFunc func(string) ([]string, error)

func (d *DAP) Run(command []string) (*Process, error) {
	d.RLock()
	if d.Process != nil {
		defer d.RUnlock()
		return d.Process, errors.New("A debug adapter is already running")
	}
	d.RUnlock()

	log.Printf("Starting debug adapter...")

	p, err := d.NewProcess(command[0], command[1:]...)
	if err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}

	d.Lock()
	d.Process = p
	d.Unlock()

	go func() {
		if err := p.Wait(); err != nil {
			log.Printf("Debug adapter exited with error: %s", err)
		} else {
			log.Print("Debug adapter exited")
		}
		d.ClearProcess()
	}()

	log.Printf("Started debug adapter")

	resp, err := p.Initialize()
	if err != nil {
		return p, fmt.Errorf("Error initializing debug adapter: %w", err)
	}
	d.Capabilities = make(map[string]bool)
	if err := json.Unmarshal(resp.Body, &d.Capabilities); err != nil {
		log.Printf("Error parsing capabilities: %s", err)
	}

	return p, nil
}

func (d *DAP) SendRequest(name string, args any) (Response, error) {
	d.Lock()
	p := d.Process
	d.Unlock()
	if p == nil {
		return Response{}, errors.New("No process running")
	}
	return p.SendRequest(name, args)
}

func (d *DAP) ClearProcess() {
	d.Lock()
	d.Process = nil
	d.Unlock()
}
