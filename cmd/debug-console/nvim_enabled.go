//go:build nvim

package main

import (
	"os"

	"github.com/dradtke/debug-console/nvim"
)

func runNvim(args []string) error {
	// nvim.Main() internally uses the flag package, so we need to provide arguments
	// by overwriting os.Args.
	os.Args = args
	return nvim.Main()
}
