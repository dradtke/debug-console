package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Dap struct{}

func (d Dap) Continue(_ struct{}, _ *struct{}) error {
	log.Print("Continuing")
	return nil
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
