package dap

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"
	"sync"
)

type Conn struct {
	cmd                *exec.Cmd
	out, err           io.ReadCloser
	in                 io.WriteCloser
	inMu               sync.Mutex
	eventHandler       func(Event)
	responseHandlers   map[int64]chan<- Response
	responseHandlersMu sync.Mutex
}

func (d *DAP) NewProcess(name string, args ...string) (*Conn, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}
	c := &Conn{
		cmd:              cmd,
		out:              stdout,
		err:              stderr,
		in:               stdin,
		eventHandler:     d.EventHandler,
		responseHandlers: make(map[int64]chan<- Response),
	}
	go c.HandleOut()
	go c.HandleErr()
	return c, nil
}

// TODO: this seems to be needed for jdtls, need to test it though.
func (d *DAP) Connect(network, addr string) (*Conn, error) {
	rawConn, err := net.Dial(network, addr)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to debug adapter at %s: %w", addr, err)
	}
	c := &Conn{
		out:              rawConn,
		in:               rawConn,
		eventHandler:     d.EventHandler,
		responseHandlers: make(map[int64]chan<- Response),
	}
	go c.HandleOut()
	// go c.HandleStderr()
	return c, nil
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
		c.in.Close()
	} else {
		log.Print("Killing debug adapter")
		if err := c.cmd.Process.Kill(); err != nil {
			log.Printf("Error killing debug adapter: %s", err)
		}
	}
}

func (c *Conn) HandleErr() {
	scanner := bufio.NewScanner(c.err)
	for scanner.Scan() {
		log.Printf("<! %s", scanner.Text())
	}
}

func (c *Conn) HandleOut() {
	var (
		scratch = make([]byte, 4096)
		buf     bytes.Buffer
	)
	for {
		_ /*rawHeaders*/, _, body, err := ReadMessage(c.out, scratch, &buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Print("dap exiting")
				return
			}
			log.Printf("dap stdout: error reading message: %s", err)
			continue
		}

		//for _, line := range strings.Split(rawHeaders, NL) {
		//	log.Printf("<< %s", line)
		//}
		//log.Println("<<")
		//for _, line := range strings.Split(body, NL) {
		//	log.Printf("<< %s", line)
		//}

		var parsed struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(body), &parsed); err != nil {
			log.Printf("dap stdout: error parsing message: %s", err)
		}

		switch parsed.Type {
		case "response":
			var resp Response
			if err := json.Unmarshal([]byte(body), &resp); err != nil {
				log.Printf("dap stdout: error parsing response: %s", err)
			}

			c.responseHandlersMu.Lock()
			ch := c.responseHandlers[resp.RequestSeq]
			delete(c.responseHandlers, resp.RequestSeq)
			c.responseHandlersMu.Unlock()

			if ch != nil {
				ch <- resp
			}

		case "event":
			var event Event
			if err := json.Unmarshal([]byte(body), &event); err != nil {
				log.Printf("dap stdout: error parsing event: %s", err)
			}
			go c.eventHandler(event)

		default:
			log.Printf("unrecognized incoming message type: %s", parsed.Type)
		}
	}
}

func (c *Conn) SendRequest(name string, args any) (Response, error) {
	req := NewRequest(name, args)
	ch := make(chan Response, 1)

	c.responseHandlersMu.Lock()
	c.responseHandlers[req.Seq] = ch
	c.responseHandlersMu.Unlock()

	if err := c.SendMessage(req); err != nil {
		c.responseHandlersMu.Lock()
		delete(c.responseHandlers, req.Seq)
		c.responseHandlersMu.Unlock()
		return Response{}, fmt.Errorf("Error sending request: %s: %w", name, err)
	}

	resp := <-ch
	if !resp.Success {
		var errorResp ErrorResponse
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
	//for _, line := range strings.Split(string(b), NL) {
	//	log.Printf(">> %s", line)
	//}
	c.inMu.Lock()
	defer c.inMu.Unlock()
	if _, err := c.in.Write(b); err != nil {
		return fmt.Errorf("Process.SendMessage: error sending message: %w", err)
	}
	return nil
}
