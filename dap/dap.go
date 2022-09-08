package dap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	"github.com/dradtke/debug-console/tmux"
	"github.com/dradtke/debug-console/types"
	"github.com/dradtke/debug-console/util"
)

type DAP struct {
	sync.RWMutex

	// Exe is the executable, used for launching the console.
	Exe string
	// Dir is where debug adapters are saved locally.
	Dir                string
	EditorEventHandler types.EventHandler
	Conn               *Conn
	Capabilities       map[string]bool
	ConsoleClient      *rpc.Client
	OutputBroadcaster  *OutputBroadcaster
	StoppedLocation    *types.StackFrame
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

	eventHandlers := []types.EventHandler{d.HandleEvent, d.EditorEventHandler}
	if conn, err = connector.Connect(eventHandlers); err != nil {
		log.Printf("Failed to start debug adapter process: %s", err)
	}

	defer func() {
		if err != nil && conn != nil {
			log.Print("Stopping existing connection")
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
	if err = d.StartConsole(); err != nil {
		return conn, fmt.Errorf("Starting console: %w", err)
	}

	resp, err := conn.Initialize()
	if err != nil {
		return conn, fmt.Errorf("Error initializing debug adapter: %w", err)
	}
	d.Capabilities = make(map[string]bool)
	if err := json.Unmarshal(resp.Body, &d.Capabilities); err != nil {
		log.Printf("Error parsing capabilities: %s", err)
	}

	if d.OutputBroadcaster, err = NewOutputBroadcaster(); err != nil {
		return conn, fmt.Errorf("Creating output broadcaster: %w", err)
	}

	return conn, nil
}

func (d *DAP) Stop() {
	d.Lock()
	defer d.Unlock()

	if d.Conn != nil {
		log.Println("Stopping running process")
		d.Conn.Stop()
	}

	d.OutputBroadcaster.Stop()

	if err := d.ConsoleClient.Call("ConsoleService.Stop", struct{}{}, nil); err != nil {
		log.Printf("Error stopping console: %s", err)
	}
}

func (d *DAP) StartConsole() error {
	consolePane, err := tmux.FindPane("console")
	if err != nil {
		return fmt.Errorf("Error finding console pane: %w", err)
	}
	if consolePane == "" {
		if err = tmux.SplitConsole(); err != nil {
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
	log.Printf("Listening for incoming rpc connections on %s", addrArg(dapListener.Addr()))

	// Grab a free port by opening a connection, and then immediately closing it.
	consoleListener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("Error opening console rpc listener: %w", err)
	}
	if err := consoleListener.Close(); err != nil {
		return fmt.Errorf("Error releasing console rpc listener: %w", err)
	}

	dapServer := rpc.NewServer()
	dapServer.Register(DAPService{d})
	go dapServer.Accept(dapListener)

	args := []string{
		d.Exe,
		"console",
		"-rpc-dap=" + addrArg(dapListener.Addr()),
		"-rpc-console=" + addrArg(consoleListener.Addr()),
	}

	if err = tmux.RunInPane(consolePane, args...); err != nil {
		return fmt.Errorf("Error running console: %w", err)
	}

	log.Printf("Connecting to console rpc on %s", addrArg(consoleListener.Addr()))
	d.ConsoleClient, err = util.TryDial(consoleListener.Addr().Network(), consoleListener.Addr().String())
	if err != nil {
		return fmt.Errorf("Error connecting to console: %w", err)
	}

	return nil
}

func (d *DAP) SendRequest(req types.Request) (types.Response, error) {
	d.Lock()
	p := d.Conn
	d.Unlock()
	if p == nil {
		return types.Response{}, errors.New("No process running")
	}
	return p.SendRequest(req)
}

func (d *DAP) ClearProcess() {
	d.Lock()
	d.Conn = nil
	d.Unlock()
}

func (d *DAP) HandleEvent(event types.Event) {
	log.Printf("Received event: %s", event.Event)

	switch event.Event {
	case "initialized":
		log.Print("Debug adapter initialized")

	case "output":
		var output types.OutputEvent
		if err := json.Unmarshal(event.Body, &output); err != nil {
			log.Printf("Error parsing output event: %s", err)
		} else if err = d.ShowOutput(output); err != nil {
			log.Printf("Error showing output: %s", err)
		}

	case "terminated":
		log.Print("Debug adapter terminated")
		d.Stop()
	}
}

func (d *DAP) HandleStopped(stopped types.StoppedEvent) (*types.StackFrame, error) {
	if err := d.ConsoleClient.Call("ConsoleService.HandleStopped", struct{}{}, nil); err != nil {
		log.Printf("Error invoking ConsoleService.HandleStopped: %s", err)
	}
	if stopped.ThreadID == nil {
		return nil, nil
	}

	resp, err := d.SendRequest(types.NewStackTraceRequest(types.StackTraceArguments{
		ThreadID: *stopped.ThreadID,
		Levels:   1,
		Format: &types.StackFrameFormat{
			Line: types.PtrBool(true),
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("Error getting stack trace: %w", err)
	}

	var body struct {
		StackFrames []types.StackFrame `json:"stackFrames"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("Error parsing stackTrace response: %w", err)
	}

	if len(body.StackFrames) == 0 {
		return nil, nil
	}

	stackFrame := body.StackFrames[0]
	d.Lock()
	d.StoppedLocation = &stackFrame
	d.Unlock()
	return &stackFrame, nil
}

func (d *DAP) ShowOutput(output types.OutputEvent) error {
	d.Lock()
	defer d.Unlock()

	if !d.OutputBroadcaster.inited {
		outputPane, err := tmux.FindOrSplitOutput()
		if err != nil {
			return fmt.Errorf("ShowOutput: %w", err)
		}

		args := []string{
			d.Exe,
			"output",
			"-addr=" + addrArg(d.OutputBroadcaster.l.Addr()),
		}

		if err = tmux.RunInPane(outputPane, args...); err != nil {
			return fmt.Errorf("ShowOutput: %w", err)
		}

		d.OutputBroadcaster.inited = true
	}

	d.OutputBroadcaster.Broadcast(output)
	return nil
}

func addrArg(addr net.Addr) string {
	return addr.Network() + " " + addr.String()
}
