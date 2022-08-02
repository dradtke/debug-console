package rpc

import (
	"fmt"
	"net"
	"net/rpc"
)

type Console struct {
	// TODO: define service methods
}

func RunConsole(listener net.Listener) {
	s := rpc.NewServer()
	s.Register(Console{})
	s.Accept(listener)
}

type ConsoleClient struct {
	c *rpc.Client
}

func NewConsoleClient(network, addr string) (ConsoleClient, error) {
	if c, err := tryDial(network, addr); err != nil {
		return ConsoleClient{}, fmt.Errorf("rpc: Error connecting to console server: %w", err)
	} else {
		return ConsoleClient{c}, nil
	}
}
