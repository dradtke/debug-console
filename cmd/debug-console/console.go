package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/dradtke/debug-console/rpc"
)

func runConsole(args []string) error {
	clearScreen()

	var (
		fs         = flag.NewFlagSet("console", flag.ExitOnError)
		rpcDap     = fs.String("rpc-dap", "", "network and address for plugin rpc")
		rpcConsole = fs.String("rpc-console", "", "network and address for console rpc")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *rpcDap == "" {
		return errors.New("-rpc-dap is required")
	}
	if *rpcConsole == "" {
		return errors.New("-rpc-console is required")
	}

	rpcDapParts := strings.Split(*rpcDap, " ")
	// log.Printf("Connecting to dap on %s %s", rpcDapParts[0], rpcDapParts[1])
	_, err := rpc.NewDapClient(rpcDapParts[0], rpcDapParts[1])
	if err != nil {
		return fmt.Errorf("Error connecting to dap server: %w", err)
	}

	rpcConsoleParts := strings.Split(*rpcConsole, " ")
	//log.Printf("Listening for incoming connections on %s %s", rpcConsoleParts[0], rpcConsoleParts[1])
	consoleListener, err := net.Listen(rpcConsoleParts[0], rpcConsoleParts[1])
	if err != nil {
		return fmt.Errorf("Error opening console rpc listener at %s: %w", *rpcConsole, err)
	}

	console, err := rpc.NewConsole()
	if err != nil {
		return fmt.Errorf("Error creating console: %w", err)
	}
	go console.Listen(consoleListener)

	if err := consoleInputLoop(console); err != nil {
		fmt.Println(err)
	}

	return nil
}

func consoleInputLoop(console rpc.Console) error {
	for {
		<-console.Stops

		input: for {
			line, err := console.Prompt.Readline()
			if err != nil {
				return err
			}
			if line == "c" || line == "continue" {
				break input
			}
			fmt.Println(line)
		}
	}
}
