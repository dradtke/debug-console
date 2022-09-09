package dap

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/dradtke/debug-console/types"
)

// Connector is an interface describing how to connect to a debug adapter.
// The two main options are to spawn a subprocess, or to connect to one that
// is already running.
type Connector interface {
	Connect(eventHandlers []types.EventHandler) (*Conn, error)
}

type Subprocess struct {
	Command       []string
	DialClientArg string
}

func (s Subprocess) Connect(eventHandlers []types.EventHandler) (*Conn, error) {
	cmd := exec.Command(s.Command[0], s.Command[1:]...)
	conn := &Conn{
		cmd:              cmd,
		eventHandlers:    eventHandlers,
		responseHandlers: make(map[int64]chan<- types.Response),
	}

	if s.DialClientArg != "" {
		// Listen for the server to connect to us
		listener, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return nil, fmt.Errorf("Error creating listener: %w", err)
		}
		defer listener.Close()
		cmd.Args = append(cmd.Args, s.DialClientArg, listener.Addr().String())
		log.Printf("Starting command: %s", strings.Join(conn.cmd.Args, " "))
		if err := conn.cmd.Start(); err != nil {
			return nil, fmt.Errorf("Error spawning debug adapter process: error starting process: %w", err)
		}
		log.Println("Waiting for connection from delve")
		c, err := listener.Accept()
		log.Println("Accepted connection (presumably from delve)")
		if err != nil {
			cmd.Process.Kill()
			return nil, fmt.Errorf("Error spawning debug adapter process: error accepting connection: %w", err)
		}
		conn.out = c
		conn.in = c
	} else {
		// Connect to the process' standard streams
		if stdout, err := cmd.StdoutPipe(); err != nil {
			return nil, fmt.Errorf("Error spawning debug adapter subprocess: error getting stdout pipe: %w", err)
		} else {
			conn.out = stdout
		}
		if stderr, err := cmd.StderrPipe(); err != nil {
			return nil, fmt.Errorf("Error spawning debug adapter subprocess: error getting stderr pipe: %w", err)
		} else {
			conn.err = stderr
		}
		if stdin, err := cmd.StdinPipe(); err != nil {
			return nil, fmt.Errorf("Error spawning debug adapter subprocess: error getting stdin pipe: %w", err)
		} else {
			conn.in = stdin
		}
		log.Printf("Starting command: %s", strings.Join(conn.cmd.Args, " "))
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("Error spawning debug adapter subprocess: error starting process: %w", err)
		}
	}

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
