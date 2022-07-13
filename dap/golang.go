package dap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// dlv is at ~/.asdf/installs/golang/1.18.3/packages/bin/dlv
// other source & docs are around there, too

func GoStart(dapDir string, eventHandler func(Event)) (*Process, error) {
	dir := filepath.Join(dapDir, "golang")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = DownloadGo(dir); err != nil {
			return nil, fmt.Errorf("Go: error downloading adapter: %w", err)
		}
	}

	p, err := NewProcess(eventHandler, "node", filepath.Join(dir, "extension/dist/debugAdapter.js"))
	if err != nil {
		return nil, fmt.Errorf("Go: error running adapter: %w", err)
	}

	return p, nil
}

func GoLaunch(filepath string, p *Process) (Response, error) {
	const dlv = "/home/damien/.asdf/installs/golang/1.18.3/packages/bin/dlv"
	log.Println("Launching Delve!")
	args := map[string]any{
		"request":     "launch",
		"program":     filepath,
		"dlvToolPath": dlv,
		"args":        []string{},
	}
	if strings.HasSuffix(filepath, "_test.go") {
		args["mode"] = "test"
	}
	return p.SendRequest("launch", args)
}

func DownloadGo(dir string) error {
	const url = "https://github.com/golang/vscode-go/releases/download/v0.34.1/go-0.34.1.vsix"
	if err := getZip(dir, url); err != nil {
		return fmt.Errorf("DownloadGo: %w", err)
	}

	return nil
}
