package dap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/dradtke/debug-console/rpc"
	"github.com/dradtke/debug-console/tmux"
)

type DAP struct {
	sync.RWMutex

	// Exe is the executable, used for launching the console.
	Exe string
	// Dir is where debug adapters are saved locally.
	Dir          string
	EventHandler func(Event)
	Conn         *Conn
	Capabilities map[string]bool
	ConsoleClient rpc.ConsoleClient
}

type DapCommandFunc func(string) ([]string, error)

func (d *DAP) Run(f func() (Connector, error)) (conn *Conn, err error) {
	d.RLock()
	if d.Conn != nil {
		defer d.RUnlock()
		return nil, errors.New("A debug adapter is already running")
	}
	d.RUnlock()

	log.Printf("Starting debug adapter...")
	connector, err := f()
	if err != nil {
		return nil, fmt.Errorf("Error creating connector: %w", err)
	}

	if conn, err = connector.Connect(d.EventHandler); err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}

	defer func() {
		if err != nil && conn != nil {
			conn.Stop()
		}
	}()

	d.Lock()
	d.Conn = conn
	d.Unlock()

	go func() {
		if err := conn.Wait(); err != nil {
			// ???: Suppress this message if the adapter was killed by Neovim exiting?
			log.Printf("Debug adapter exited with error: %s", err)
		} else {
			log.Print("Debug adapter exited")
		}
		d.ClearProcess()
	}()

	log.Printf("Started debug adapter")
	if err = d.StartConsole(1); err != nil {
		return nil, fmt.Errorf("Starting starting console: %w", err)
	}

	resp, err := conn.Initialize()
	if err != nil {
		return conn, fmt.Errorf("Error initializing debug adapter: %w", err)
	}
	d.Capabilities = make(map[string]bool)
	if err := json.Unmarshal(resp.Body, &d.Capabilities); err != nil {
		log.Printf("Error parsing capabilities: %s", err)
	}

	return conn, nil
}

func (d *DAP) StartConsole(pane int) error {
	consolePane, err := tmux.FindPane("console")
	if err != nil {
		return fmt.Errorf("Error finding console pane: %w", err)
	}
	if consolePane == "" {
		if err = tmux.Split(); err != nil {
			return fmt.Errorf("Error splitting tmux panes: %w", err)
		}
		if consolePane, err = tmux.FindPane("console"); err != nil {
			return fmt.Errorf("Error finding console pane: %w", err)
		}
	}

	dapListener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("Error opening DAP rpc listener: %w", err)
	}

	// Grab a free port by opening a connection, and then immediately closing it.
	consoleListener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("Error opening console rpc listener: %w", err)
	}
	if err := consoleListener.Close(); err != nil {
		return fmt.Errorf("Error releasing console rpc listener: %w", err)
	}

	go rpc.RunDap(dapListener)

	rpcAddr := func(addr net.Addr) string {
		return addr.Network() + " " + addr.String()
	}

	args := []string{
		d.Exe,
		"console",
		"-rpc-dap=" + rpcAddr(dapListener.Addr()),
		"-rpc-console=" + rpcAddr(consoleListener.Addr()),
	}

	if err = tmux.RunInPane(consolePane, args...); err != nil {
		return fmt.Errorf("Error running console: %w", err)
	}

	d.ConsoleClient, err = rpc.NewConsoleClient(consoleListener.Addr().Network(), consoleListener.Addr().String())
	if err != nil {
		return fmt.Errorf("Error connecting to console: %w", err)
	}

	return nil
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
