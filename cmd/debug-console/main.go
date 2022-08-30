package main

import (
	"log"
	"os"
)

func main() {
	log.Print("Starting debug console...")

	if len(os.Args) < 2 {
		log.Fatal("expected at least one argument")
	}
	cmd := os.Args[1]
	args := os.Args[1:]

	funcs := map[string]func([]string) error{
		"console": runConsole,
		"nvim":    runNvim,
		"output": runOutput,
	}

	f, ok := funcs[cmd]
	if !ok {
		log.Fatalf("Unrecognized command: %s", cmd)
	}

	if err := f(args); err != nil {
		log.Fatalf("Error running command %s: %s", cmd, err)
	}

	log.Print("Exiting")
}
