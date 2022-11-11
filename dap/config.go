package dap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/dradtke/debug-console/types"
)

type ConfigMap map[string]Config

type Configs struct {
	// default by filetype + name
	Defaults map[string]ConfigMap `msgpack:"defaults"`

	// user configs by name
	User ConfigMap `msgpack:"user"`
}

func (c Configs) Get(filetype, name string) (Config, error) {
	if config, ok := c.User[name]; ok {
		return config, nil
	}
	if filetypeConfigs := c.Defaults[filetype]; filetypeConfigs != nil {
		if config, ok := filetypeConfigs[name]; ok {
			return config, nil
		}
	}
	return Config{}, fmt.Errorf("no configuration found for name: %s", name)
}

func (c Configs) Available(filetype string) []string {
	var result []string
	for name := range c.User {
		result = append(result, name)
	}
	if filetypeConfigs := c.Defaults[filetype]; filetypeConfigs != nil {
		for name := range filetypeConfigs {
			result = append(result, name)
		}
	}
	return result
}

type Config struct {
	Run    RunSpec `msgpack:"run"`
	Launch string  `msgpack:"launch"`
}

// RunSpec defines how the DAP server should be started or connected to.
type RunSpec struct {
	Type    string   `msgpack:"type"`    // subprocess, lsp command (jdtls), etc.
	Command []string `msgpack:"command"` // populated for subprocess
	// DialClient tells the client to open a listener for the server to connect
	// to. The server's command should include the value "${CLIENT_ADDR}",
	// which will be replaced with the client's address.
	DialClient bool `msgpack:"dialClient"`
}

func (r RunSpec) Run(eventHandlers []types.EventHandler) (*Conn, error) {
	switch r.Type {
	case "subprocess":
		return r.runSubprocess(eventHandlers)
	default:
		return nil, fmt.Errorf("unknown run type: %s", r.Type)
	}
}

func (r RunSpec) runSubprocess(eventHandlers []types.EventHandler) (*Conn, error) {
	cmd := exec.Command(r.Command[0], r.Command[1:]...)
	conn := &Conn{
		cmd:                  cmd,
		eventHandlers:        eventHandlers,
		responseHandlers:     make(map[int64]chan<- types.Response),
		initializedEventSeen: make(chan struct{}),
	}

	if err := conn.pipeStreams(); err != nil {
		return nil, err
	}

	if r.DialClient {
		go broadcastAsOutput("stdout", conn.out, eventHandlers)
		go broadcastAsOutput("stderr", conn.err, eventHandlers)
		if err := r.runConnectingSubprocess(conn); err != nil {
			return nil, err
		}
	} else {
		log.Printf("Starting debug adapter with command: %s", strings.Join(conn.cmd.Args, " "))
		if err := conn.cmd.Start(); err != nil {
			return nil, err
		}
		go conn.HandleOut()
		go conn.HandleErr()
	}

	return conn, nil
}

func (r RunSpec) runConnectingSubprocess(conn *Conn) error {
	// Listen for the server to connect to us
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("Error creating listener: %w", err)
	}
	for i := range conn.cmd.Args {
		conn.cmd.Args[i] = strings.Replace(conn.cmd.Args[i], "${CLIENT_ADDR}", listener.Addr().String(), -1)
	}
	log.Printf("Starting debug adapter with command: %s", strings.Join(conn.cmd.Args, " "))
	if err := conn.cmd.Start(); err != nil {
		return err
	}

	conn.inMu.Lock()

	go func() {
		defer listener.Close()
		defer conn.inMu.Unlock()
		log.Print("Waiting for connection from subprocess")
		c, err := listener.Accept()
		if err != nil {
			log.Printf("Error waiting for connection: %s", err)
			conn.cmd.Process.Kill()
		} else {
			log.Print("Got connection from subprocess")
			conn.out = c
			conn.in = c
			go conn.HandleOut()
			go conn.HandleErr()
		}
	}()
	return nil
}

func broadcastAsOutput(category string, r io.Reader, eventHandlers []types.EventHandler) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		body, err := json.Marshal(types.OutputEvent{
			Category: category,
			Output: line + "\n",
		})
		if err != nil {
			log.Println("broadcastAsOutput: Error marshaling output event: " + err.Error())
			return
		}
		event := types.Event{
			Event: "output",
			Body: json.RawMessage(body),
		}
		for _, eventHandler := range eventHandlers {
			eventHandler(event)
		}
	}
}
