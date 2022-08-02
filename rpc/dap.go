package rpc

import (
	"fmt"
	"net"
	"net/rpc"
)

type Dap struct {
	// TODO: define service methods
}

func RunDap(listener net.Listener) {
	s := rpc.NewServer()
	s.Register(Dap{})
	s.Accept(listener)
}

type DapClient struct {
	c *rpc.Client
}

func NewDapClient(network, addr string) (DapClient, error) {
	if c, err := tryDial(network, addr); err != nil {
		return DapClient{}, fmt.Errorf("rpc: Error connecting to dap server: %w", err)
	} else {
		return DapClient{c}, nil
	}
}
