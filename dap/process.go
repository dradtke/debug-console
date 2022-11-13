package dap

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/dradtke/debug-console/tmux"
	"github.com/dradtke/debug-console/types"
	"github.com/dradtke/debug-console/util"
)

const VerboseLogging = true

type Conn struct {
	cmd                  *exec.Cmd
	out, err             io.ReadCloser
	outDone              chan struct{}
	in                   io.WriteCloser
	inMu                 sync.Mutex
	eventHandlers        []types.EventHandler
	responseHandlers     map[int64]chan<- types.Response
	responseHandlersMu   sync.Mutex
	initializedEventSeen chan struct{}
	seeInitializeEvent   sync.Once
}

func (c *Conn) Wait() error {
	if c.cmd == nil {
		return nil
	}
	return c.cmd.Wait()
}

func (c *Conn) Stop() {
	if c.cmd == nil {
		// ???: Is this enough to tell the connection to stop?
		c.out.Close()
		c.err.Close()
		c.in.Close()
	} else {
		log.Print("Killing debug adapter")
		if err := c.cmd.Process.Kill(); err != nil {
			log.Printf("Error killing debug adapter: %s", err)
		}
	}
}

func (c *Conn) HandleErr() {
	if c.err == nil {
		return
	}
	scanner := bufio.NewScanner(c.err)
	for scanner.Scan() {
		log.Printf("<! %s", scanner.Text())
	}
}

func (c *Conn) HandleOut() {
	defer func() {
		log.Println("Done reading stdout")
		close(c.outDone)
	}()
	if c.out == nil {
		log.Printf("No output stream to read!")
		return
	}
	var (
		scratch = make([]byte, 4096)
		buf     bytes.Buffer
	)
	for {
		_ /*headers*/, rawHeaders, body, err := ReadMessage(c.out, scratch, &buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Print("dap exiting")
				return
			}
			log.Printf("dap stdout: error reading message: %s", err)
			continue
		}

		/*
			var buf bytes.Buffer
			if err := json.Indent(&buf, []byte(body), "", "  "); err != nil {
				panic(err)
			}
			log.Println(buf.String())
		*/

		if VerboseLogging {
			for _, line := range strings.Split(rawHeaders, NL) {
				log.Printf("<< %s", line)
			}
			log.Println("<<")
			for _, line := range strings.Split(body, NL) {
				log.Printf("<< %s", line)
			}
		}

		var parsed struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(body), &parsed); err != nil {
			log.Printf("dap stdout: error parsing message: %s", err)
		}

		switch parsed.Type {
		case "response":
			var resp types.Response
			if err := json.Unmarshal([]byte(body), &resp); err != nil {
				log.Printf("dap stdout: error parsing response: %s", err)
			} else {
				c.responseHandlersMu.Lock()
				ch := c.responseHandlers[resp.RequestSeq]
				delete(c.responseHandlers, resp.RequestSeq)
				c.responseHandlersMu.Unlock()

				if ch != nil {
					ch <- resp
				}
			}

		case "event":
			var event types.Event
			if err := json.Unmarshal([]byte(body), &event); err != nil {
				log.Printf("dap stdout: error parsing event: %s", err)
			} else {
				for _, f := range c.eventHandlers {
					f(event)
				}
			}

		case "request":
			var req types.ReverseRequest
			if err := json.Unmarshal([]byte(body), &req); err != nil {
				log.Printf("dap stdout: error parsing reverse request: %s", err)
			} else {
				go c.HandleReverseRequest(req)
			}

		default:
			log.Printf("unrecognized incoming message type: %s", parsed.Type)
		}
	}
}

func (c *Conn) HandleReverseRequest(req types.ReverseRequest) {
	defer util.Recover()

	switch req.Command {
	case "runInTerminal":
		pane, err := tmux.FindOrSplitRunInTerminal()
		if err != nil {
			log.Printf("Failed to find or split run-in-terminal tmux pane: %s", err)
			return
		}
		if cwd := req.Arguments["cwd"]; cwd != nil {
			// TODO: use
		}
		var args []string
		for _, v := range req.Arguments["args"].([]any) {
			args = append(args, v.(string))
		}
		// TODO: this isn't connecting to the debug adapter, for some reason...
		if err = tmux.RunInPaneNoQuote(pane, args...); err != nil {
			log.Printf("Failed to run in run-in-terminal tmux pane: %s", err)
		}

	default:
		log.Printf("Unknown reverse request command: %s", req.Command)
	}
}

func (c *Conn) SendRequest(req types.Request) (types.Response, error) {
	ch := make(chan types.Response, 1)

	c.responseHandlersMu.Lock()
	c.responseHandlers[req.Seq()] = ch
	c.responseHandlersMu.Unlock()

	if err := c.SendMessage(req); err != nil {
		c.responseHandlersMu.Lock()
		delete(c.responseHandlers, req.Seq())
		c.responseHandlersMu.Unlock()
		return types.Response{}, fmt.Errorf("Error sending request: %s: %w", req.Command(), err)
	}

	resp := <-ch
	if !resp.Success {
		var errorResp types.ErrorResponse
		if err := json.Unmarshal(resp.Body, &errorResp); err != nil {
			return resp, fmt.Errorf("Error unmarshaling error response: %w", err)
		}
		return resp, errorResp
	}
	return resp, nil
}

func (c *Conn) SendMessage(msg any) error {
	b, err := Message(msg)
	if err != nil {
		return fmt.Errorf("Process.SendMessage: error building message: %w", err)
	}

	if VerboseLogging {
		for _, line := range strings.Split(string(b), NL) {
			log.Printf(">> %s", line)
		}
	}

	c.inMu.Lock()
	defer c.inMu.Unlock()
	if _, err := c.in.Write(b); err != nil {
		return fmt.Errorf("Process.SendMessage: error sending message: %w", err)
	}
	return nil
}

func (c *Conn) InitializedEventSeen() <-chan struct{} {
	return c.initializedEventSeen
}

func (c *Conn) pipeStreams() error {
	// Connect to the process' standard streams
	if stdout, err := c.cmd.StdoutPipe(); err != nil {
		return err
	} else {
		c.out = stdout
	}
	if stderr, err := c.cmd.StderrPipe(); err != nil {
		return err
	} else {
		c.err = stderr
	}
	if stdin, err := c.cmd.StdinPipe(); err != nil {
		return err
	} else {
		c.in = stdin
	}
	return nil
}
