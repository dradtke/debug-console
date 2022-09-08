package dap

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/dradtke/debug-console/types"
)

// Connector is an interface describing how to connect to a debug adapter.
// The two main options are to spawn a subprocess, or to connect to one that
// is already running.
type Connector interface {
	Connect(eventHandlers []types.EventHandler) (*Conn, error)
}

type Subprocess []string

func (s Subprocess) Connect(eventHandlers []types.EventHandler) (*Conn, error) {
	cmd := exec.Command(s[0], s[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Error spawning debug adapter subprocess: error getting stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("Error spawning debug adapter subprocess: error getting stderr pipe: %w", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("Error spawning debug adapter subprocess: error getting stdin pipe: %w", err)
	}
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("Error spawning debug adapter subprocess: error starting process: %w", err)
	}
	conn := &Conn{
		cmd:              cmd,
		out:              stdout,
		err:              stderr,
		in:               stdin,
		eventHandlers:    eventHandlers,
		responseHandlers: make(map[int64]chan<- types.Response),
	}
	// TODO: how to reference the *DAP here?
	go conn.HandleOut()
	go conn.HandleErr()
	return conn, nil
}

type Connection struct {
	Network, Address string
}

func (c Connection) Connect(eventHandlers []types.EventHandler) (*Conn, error) {
	rawConn, err := net.Dial(c.Network, c.Address)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to debug adapter at %s: %w", c.Address, err)
	}
	conn := &Conn{
		out:              rawConn,
		in:               rawConn,
		eventHandlers:    eventHandlers,
		responseHandlers: make(map[int64]chan<- types.Response),
	}
	go conn.HandleOut()
	// go c.HandleStderr()
	return conn, nil
}
