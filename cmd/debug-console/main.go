package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("expected at least one argument")
	}
	cmd := os.Args[1]
	args := os.Args[1:]

	funcs := map[string]func([]string) error{
		"console": runConsole,
		"nvim":    runNvim,
		"output":  runOutput,
	}

	f, ok := funcs[cmd]
	if !ok {
		log.Fatalf("Unrecognized command: %s", cmd)
	}

	if err := f(args); err != nil {
		log.Fatalf("Error running command %s: %s", cmd, err)
	}
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		// TODO
	default:
		log.Printf("don't know how to clear screen for os: %s", runtime.GOOS)
	}
}
