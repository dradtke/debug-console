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

func GoStart(dapDir string, handlers Handlers) (*Process, OnInitializedFunc, error) {
	dir := filepath.Join(dapDir, "golang")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = DownloadGo(dir); err != nil {
			return nil, nil, fmt.Errorf("Go: error downloading adapter: %w", err)
		}
	}

	p, err := NewProcess(handlers, "node", filepath.Join(dir, "extension/dist/debugAdapter.js"))
	if err != nil {
		return nil, nil, fmt.Errorf("Go: error running adapter: %w", err)
	}

	return p, GoInitialized, nil
}

func GoInitialized(filepath string, p *Process) {
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
	p.SendRequest("launch", args)
}

func DownloadGo(dir string) error {
	const url = "https://github.com/golang/vscode-go/releases/download/v0.34.1/go-0.34.1.vsix"
	if err := getZip(dir, url); err != nil {
		return fmt.Errorf("DownloadGo: %w", err)
	}

	return nil
}
