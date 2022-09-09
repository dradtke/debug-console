package dap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dradtke/debug-console/types"
)

// dlv is at ~/.asdf/installs/golang/1.18.3/packages/bin/dlv
// other source & docs are around there, too

func GoConnector(dapDir string) func() (Connector, error) {
	return func() (Connector, error) {
		dir := filepath.Join(dapDir, "golang")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err = DownloadGo(dir); err != nil {
				return nil, fmt.Errorf("Go: error downloading adapter: %w", err)
			}
		}

		return Subprocess{
			Command: []string{
				"node",
				filepath.Join(dir, "extension/dist/debugAdapter.js"),
			},
		}, nil
		// TODO: "dlv dap" doesn't seem to work correctly yet
		/*
			return Subprocess{
				Command:       []string{"dlv", "dap"},
				DialClientArg: "--client-addr",
			}, nil
		*/
	}
}

func (d *DAP) GoLaunch(filepath string) (types.Response, error) {
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
	return d.SendRequest(types.NewLaunchRequest(args))
}

func DownloadGo(dir string) error {
	const url = "https://github.com/golang/vscode-go/releases/download/v0.34.1/go-0.34.1.vsix"
	if err := getZip(dir, url); err != nil {
		return fmt.Errorf("DownloadGo: %w", err)
	}

	return nil
}
