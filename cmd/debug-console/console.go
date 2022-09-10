package main

import (
	"encoding/json"
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

	console, err := console.NewConsole(dapClient)
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
	var (
		multiline bool
		lines     []string
	)

	// TODO: remove the non-interactivity
	// For some reason doing that breaks the plugin?
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
					// Give the debug adapter a chance to terminate gracefully.
					if err := dapClient.Call("DAPService.Terminate", struct{}{}, nil); err != nil {
						// That failed, so just disconnect.
						if err = dapClient.Call("DAPService.Disconnect", struct{}{}, nil); err != nil {
							log.Printf("Error disconnecting from debug adapter: %s", err)
						}
					}
					return nil
				}
				return err
			}
			if multiline {
				lines = append(lines, line)
				if line != "" {
					c.Prompt.SetPrompt("                *> ")
					continue input
				} else {
					line = strings.Join(lines, "\n")
					lines = nil
				}
			}
			// TODO: how to handle multiline-switching command?
			keepLooping, toggleMultiline := handleCommand(line, dapClient)
			if !keepLooping {
				break input
			}
			if toggleMultiline {
				multiline = !multiline
			}
			if multiline {
				c.Prompt.SetPrompt("debug (multiline)> ")
			} else {
				c.Prompt.SetPrompt("debug> ")
			}
		}
	}
}

func handleCommand(line string, dapClient *rpc.Client) (keepLooping, toggleMultiline bool) {
	words := strings.Split(line, " ")
	switch words[0] {
	case "?", "h", "help":
		help()
		return true, false

	case "ml", "multiline":
		return true, true

	case "caps", "capabilities":
		var capabilities types.Capabilities
		if err := dapClient.Call("DAPService.Capabilities", struct{}{}, &capabilities); err != nil {
			log.Printf("Error getting capabilities: %s", err)
		} else if b, err := json.MarshalIndent(capabilities, "", "  "); err != nil {
			log.Printf("Error formatting capabilities: %s", err)
		} else {
			fmt.Println(string(b))
		}
		return true, false

	case "threads":
		var threads []types.Thread
		if err := dapClient.Call("DAPService.Threads", struct{}{}, &threads); err != nil {
			log.Printf("Error calling threads: %s", err)
		} else {
			for _, thread := range threads {
				fmt.Printf("[%d] %s\n", thread.ID, thread.Name)
			}
		}
		return true, false

	case "c", "cont", "continue":
		if err := dapClient.Call("DAPService.Continue", struct{}{}, nil); err != nil {
			log.Printf("Error calling continue: %s", err)
		}
		return false, false

	case "step":
		if len(words) < 2 {
			fmt.Println("Must specify 'in', 'out', or 'back'")
			return true, false
		}
		switch stepDir := words[1]; stepDir {
		case "in":
			if err := dapClient.Call("DAPService.StepIn", struct{}{}, nil); err != nil {
				log.Printf("Error calling stepIn: %s", err)
			}
		case "out":
			if err := dapClient.Call("DAPService.StepOut", struct{}{}, nil); err != nil {
				log.Printf("Error calling stepOut: %s", err)
			}
		case "back":
			// this will need more testing....
			if err := dapClient.Call("DAPService.StepBack", struct{}{}, nil); err != nil {
				log.Printf("Error calling stepBack: %s", err)
			}
			return true, false
		default:
			fmt.Printf("Unknown step direction: %s\n", stepDir)
			return true, false
		}
		return false, false

	case "n", "next":
		var granularity string
		if len(words) > 1 {
			granularity = words[1]
		}
		if err := dapClient.Call("DAPService.Next", granularity, nil); err != nil {
			log.Printf("Error calling next: %s", err)
		}
		return false, false

	case "e", "eval", "evaluate":
		evaluate(dapClient, strings.Join(words[1:], " "))
		return true, false

	default:
		evaluate(dapClient, line)
		return true, false
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

func help() {
	fmt.Print(`Available commands:

  ?, h, help                              Show this help
  ml, multiline                           Toggle multiline mode
  caps, capabilities                      View the DAP server's capabilities
  c, cont, continue                       Continue execution
  n, next [statement|line|instruction]    Next statement, line, or instruction (default: statement)
  step (in, out, back)                    Step in, out, or back
  e, eval, evaluate [statement]           Evaluate a statement
  threads                                 Show running threads

Unrecognized commands will be evaluated as a statement.

`)
}
