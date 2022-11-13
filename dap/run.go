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

type RunArgs struct {
	Type string `msgpack:"type"`
	// Command is expected when Type is 'subprocess'
	Command    []string `msgpack:"command"`
	DialClient bool     `msgpack:"dialClient"`
	// Address is expected when Type is 'remote'
	Address string `msgpack:"address"`
}

func (r RunArgs) Run(eventHandlers []types.EventHandler) (*Conn, error) {
	conn := &Conn{
		eventHandlers:        eventHandlers,
		responseHandlers:     make(map[int64]chan<- types.Response),
		initializedEventSeen: make(chan struct{}),
		outDone:              make(chan struct{}),
	}
	switch r.Type {
	case "subprocess":
		return conn, r.runSubprocess(conn)
	case "remote":
		return conn, r.runRemote(conn)
	default:
		return nil, fmt.Errorf("unknown run type: %s", r.Type)
	}
}

func (r RunArgs) runSubprocess(conn *Conn) error {
	conn.cmd = exec.Command(r.Command[0], r.Command[1:]...)

	if err := conn.pipeStreams(); err != nil {
		return err
	}

	if r.DialClient {
		go broadcastAsOutput("stdout", conn.out, conn.eventHandlers)
		go broadcastAsOutput("stderr", conn.err, conn.eventHandlers)
		if err := r.runConnectingSubprocess(conn); err != nil {
			return err
		}
	} else {
		log.Printf("Starting debug adapter with command: %s", strings.Join(conn.cmd.Args, " "))
		if err := conn.cmd.Start(); err != nil {
			return err
		}
		go conn.HandleOut()
		go conn.HandleErr()
	}

	return nil
}

func (r RunArgs) runConnectingSubprocess(conn *Conn) error {
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

func (r RunArgs) runRemote(conn *Conn) error {
	log.Printf("Dialing %s", r.Address)
	c, err := net.Dial("tcp", r.Address)
	if err != nil {
		return err
	}
	log.Println("Connected!")
	conn.inMu.Lock()
	defer conn.inMu.Unlock()
	conn.out = c
	conn.in = c
	go conn.HandleOut()
	return nil
}

func broadcastAsOutput(category string, r io.Reader, eventHandlers []types.EventHandler) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		body, err := json.Marshal(types.OutputEvent{
			Category: category,
			Output:   line + "\n",
		})
		if err != nil {
			log.Println("broadcastAsOutput: Error marshaling output event: " + err.Error())
			return
		}
		event := types.Event{
			Event: "output",
			Body:  json.RawMessage(body),
		}
		for _, eventHandler := range eventHandlers {
			eventHandler(event)
		}
	}
}
