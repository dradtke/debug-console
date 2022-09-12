package dap

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/dradtke/debug-console/types"
)

// Map from filetype to Config
type ConfigMap map[string]Config

type Config struct {
	RunField ConfigRun `msgpack:"run"`
	// TODO: ability to specify default launch configuration?
	LaunchArgFuncs map[string]string `msgpack:"launch"`
}

type ConfigRun struct {
	Type    string   `msgpack:"type"`    // subprocess, lsp command (jdtls), etc.
	Command []string `msgpack:"command"` // populated for subprocess

	// Used for 'dlv dap' and anything that behaves similarly.
	DialClientArg string `msgpack:"dialClientArg"`
}

func (r ConfigRun) Run(eventHandlers []types.EventHandler) (*Conn, error) {
	switch r.Type {
	case "subprocess":
		return r.runSubprocess(eventHandlers)
	default:
		return nil, fmt.Errorf("unknown run type: %s", r.Type)
	}
}

func (r ConfigRun) runSubprocess(eventHandlers []types.EventHandler) (*Conn, error) {
	cmd := exec.Command(r.Command[0], r.Command[1:]...)
	conn := &Conn{
		cmd:              cmd,
		eventHandlers:    eventHandlers,
		responseHandlers: make(map[int64]chan<- types.Response),
	}

	if r.DialClientArg != "" {
		// Listen for the server to connect to us
		listener, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return nil, fmt.Errorf("Error creating listener: %w", err)
		}
		defer listener.Close()
		cmd.Args = append(cmd.Args, r.DialClientArg, listener.Addr().String())
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
