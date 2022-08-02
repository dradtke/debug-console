//go:build nvim

package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dradtke/debug-console/nvim"
)

func runNvim(args []string) error {
	// nvim.Main() internally uses the flag package, so we need to provide arguments
	// by overwriting os.Args.
	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Printf("Error calculating absolute path of executable: %s", err)
		exe = os.Args[0]
	}
	os.Args = args
	return nvim.Main(exe)
}
