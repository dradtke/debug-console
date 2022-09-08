package main

import (
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/dradtke/debug-console/types"
)

func runOutput(args []string) error {
	clearScreen()

	var (
		fs         = flag.NewFlagSet("output", flag.ExitOnError)
		addr     = fs.String("addr", "", "network and address to connect to for output")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *addr == "" {
		return errors.New("-addr is required")
	}
	addrParts := strings.Split(*addr, " ")

	c, err := net.Dial(addrParts[0], addrParts[1])
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(c)
	var output types.OutputEvent

	for {
		if err := dec.Decode(&output); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		fmt.Printf("[%s] %s", output.Category, output.Output)
	}
}
