package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
)

func runOutput(args []string) error {
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

	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return scanner.Err()
}
