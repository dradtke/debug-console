package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"strings"

	"github.com/chzyer/readline"

	"github.com/dradtke/debug-console/console"
	"github.com/dradtke/debug-console/types"
	"github.com/dradtke/debug-console/util"
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
	dapClient, err := util.TryDial(rpcDapParts[0], rpcDapParts[1])
	if err != nil {
		return fmt.Errorf("Error connecting to dap server: %w", err)
	}

	rpcConsoleParts := strings.Split(*rpcConsole, " ")
	//log.Printf("Listening for incoming connections on %s %s", rpcConsoleParts[0], rpcConsoleParts[1])
	consoleListener, err := net.Listen(rpcConsoleParts[0], rpcConsoleParts[1])
	if err != nil {
		return fmt.Errorf("Error opening console rpc listener at %s: %w", *rpcConsole, err)
	}

	console, err := console.NewConsole()
	if err != nil {
		return fmt.Errorf("Error creating console: %w", err)
	}

	s := rpc.NewServer()
	s.Register(console)
	go s.Accept(consoleListener)

	if err := consoleInputLoop(console, dapClient); err != nil {
		fmt.Println(err)
	}

	return nil
}

func consoleInputLoop(c console.ConsoleService, dapClient *rpc.Client) error {
	for {
		// TODO: see if the program is running or not?
		fmt.Println("Running...")
		<-c.Stops

	input:
		for {
			line, err := c.Prompt.Readline()
			if err != nil {
				if errors.Is(err, readline.ErrInterrupt) {
					fmt.Println("Use Ctrl-D to quit")
					continue input
				}
				if errors.Is(err, io.EOF) {
					fmt.Println("Quitting...")
					if err := dapClient.Call("DAPService.Terminate", struct{}{}, nil); err != nil {
						log.Printf("Error terminating adapter: %s\n", err)
					}
					return nil
				}
				return err
			}
			keepLooping := handleCommand(line, dapClient)
			if !keepLooping {
				break input
			}
		}
	}
}

func handleCommand(line string, dapClient *rpc.Client) (keepLooping bool) {
	words := strings.Split(line, " ")
	switch words[0] {
	case "?", "h", "help":
		fmt.Println("TODO: put help here")
		return true

	case "threads":
		var threads []types.Thread
		if err := dapClient.Call("DAPService.Threads", struct{}{}, &threads); err != nil {
			log.Printf("Error calling threads: %s", err)
		} else {
			for _, thread := range threads {
				fmt.Printf("[%d] %s\n", thread.ID, thread.Name)
			}
		}
		return true

	case "c", "continue":
		if err := dapClient.Call("DAPService.Continue", struct{}{}, nil); err != nil {
			log.Printf("Error calling continue: %s", err)
		}
		return false

	case "e", "eval", "evaluate":
		evaluate(dapClient, strings.Join(words[1:], " "))
		return true

	default:
		evaluate(dapClient, line)
		return true
	}
}

func evaluate(dapClient *rpc.Client, expression string) {
	var result string
	if err := dapClient.Call("DAPService.Evaluate", types.EvaluateArguments{
		Expression: expression,
		Context:    "repl",
	}, &result); err != nil {
		log.Print(err)
	} else {
		fmt.Println(result)
	}
}
