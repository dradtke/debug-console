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
	"sync"
)

type Process struct {
	cmd                *exec.Cmd
	stdout, stderr     io.ReadCloser
	stdin              io.WriteCloser
	stdinMu            sync.Mutex
	eventHandler       func(Event)
	responseHandlers   map[int64]chan<- Response
	responseHandlersMu sync.Mutex
}

func (d *DAP) NewProcess(name string, args ...string) (*Process, error) {
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
	p := &Process{
		cmd:          cmd,
		stdout:       stdout,
		stderr:       stderr,
		stdin:        stdin,
		eventHandler: d.EventHandler,
		responseHandlers: make(map[int64]chan<- Response),
	}
	go p.HandleStdout()
	go p.HandleStderr()
	return p, nil
}

func (p *Process) Wait() error {
	return p.cmd.Wait()
}

func (p *Process) Stop() {
	log.Print("Killing debug adapter")
	if err := p.cmd.Process.Kill(); err != nil {
		log.Printf("Error killing debug adapter: %s", err)
	}
}

func (p *Process) HandleStderr() {
	scanner := bufio.NewScanner(p.stderr)
	for scanner.Scan() {
		log.Printf("<! %s", scanner.Text())
	}
}

func (p *Process) HandleStdout() {
	var (
		scratch = make([]byte, 4096)
		buf     bytes.Buffer
	)
	for {
		_ /*rawHeaders*/, _, body, err := ReadMessage(p.stdout, scratch, &buf)
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

			p.responseHandlersMu.Lock()
			ch := p.responseHandlers[resp.RequestSeq]
			delete(p.responseHandlers, resp.RequestSeq)
			p.responseHandlersMu.Unlock()

			if ch != nil {
				ch <- resp
			}

		case "event":
			var event Event
			if err := json.Unmarshal([]byte(body), &event); err != nil {
				log.Printf("dap stdout: error parsing event: %s", err)
			}
			go p.eventHandler(event)

		default:
			log.Printf("unrecognized incoming message type: %s", parsed.Type)
		}
	}
}

func (p *Process) SendRequest(name string, args any) (Response, error) {
	req := NewRequest(name, args)
	ch := make(chan Response, 1)

	p.responseHandlersMu.Lock()
	p.responseHandlers[req.Seq] = ch
	p.responseHandlersMu.Unlock()

	if err := p.SendMessage(req); err != nil {
		p.responseHandlersMu.Lock()
		delete(p.responseHandlers, req.Seq)
		p.responseHandlersMu.Unlock()
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

func (p *Process) SendMessage(msg any) error {
	b, err := Message(msg)
	if err != nil {
		return fmt.Errorf("Process.SendMessage: error building message: %w", err)
	}
	//for _, line := range strings.Split(string(b), NL) {
	//	log.Printf(">> %s", line)
	//}
	p.stdinMu.Lock()
	defer p.stdinMu.Unlock()
	if _, err := p.stdin.Write(b); err != nil {
		return fmt.Errorf("Process.SendMessage: error sending message: %w", err)
	}
	return nil
}

func (p *Process) Close() error {
	return p.cmd.Wait()
}
