package dap

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const NL = "\r\n"

type OnInitializedFunc func(string, *Process)

type Handlers struct {
	Response func(*Process, Response)
	Event    func(Event)
}

type Process struct {
	cmd            *exec.Cmd
	stdout, stderr io.ReadCloser
	stdin          io.WriteCloser
	handlers       Handlers
}

func NewProcess(handlers Handlers, name string, args ...string) (*Process, error) {
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
		cmd:      cmd,
		stdout:   stdout,
		stderr:   stderr,
		stdin:    stdin,
		handlers: handlers,
	}
	go p.HandleStdout()
	go p.HandleStderr()
	return p, nil
}

func (p *Process) Wait() error {
	return p.cmd.Wait()
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
		_, rawHeaders, body, err := ReadMessage(p.stdout, scratch, &buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Print("dap exiting")
				return
			}
			log.Printf("dap stdout: error reading message: %s", err)
			continue
		}

		for _, line := range strings.Split(rawHeaders, NL) {
			log.Printf("<< %s", line)
		}
		log.Println("<<")
		for _, line := range strings.Split(body, NL) {
			log.Printf("<< %s", line)
		}

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
			p.handlers.Response(p, resp)

		case "event":
			var event Event
			if err := json.Unmarshal([]byte(body), &event); err != nil {
				log.Printf("dap stdout: error parsing event: %s", err)
			}
			p.handlers.Event(event)

		default:
			log.Printf("unrecognized incoming message type: %s", parsed.Type)
		}
	}
}

func (p *Process) SendRequest(name string, args any) {
	if err := p.SendMessage(NewRequest(name, args)); err != nil {
		log.Printf("Error sending request: %s: %s", name, err)
	}
}

func (p *Process) SendMessage(msg any) error {
	b, err := Message(msg)
	if err != nil {
		return fmt.Errorf("Process.SendMessage: error building message: %w", err)
	}
	for _, line := range strings.Split(string(b), NL) {
		log.Printf(">> %s", line)
	}
	if _, err := p.stdin.Write(b); err != nil {
		return fmt.Errorf("Process.SendMessage: error sending message: %w", err)
	}
	return nil
}

func (p *Process) Initialize() {
	// TODO: create a struct for this?
	p.SendRequest("initialize", map[string]any{
		"adapterID":                    "debug-console",
		"pathFormat":                   "path",
		"linesStartAt1":                true,
		"columnsStartAt1":              true,
		"supportsRunInTerminalRequest": true,
	})
}

func (p *Process) Close() error {
	return p.cmd.Wait()
}

func getZip(dst, src string) error {
	log.Printf("Downloading zip archive from %s to %s", src, dst)
	resp, err := http.Get(src)
	if err != nil {
		return fmt.Errorf("getZip: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("getZip: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return fmt.Errorf("getZip: %w", err)
	}

	// TODO: need to save these files to dapDir/name, preserving their structure
	for _, f := range r.File {
		r, err := f.Open()
		if err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
		fileDest := filepath.Join(dst, f.Name)
		if err := os.MkdirAll(filepath.Dir(fileDest), 0755); err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
		f, err := os.OpenFile(fileDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
		if _, err := io.Copy(f, r); err != nil {
			return fmt.Errorf("getZip: %w", err)
		}
	}
	log.Print("Download complete")

	return nil
}
