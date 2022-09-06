package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/dradtke/debug-console/types"
)

type Dap struct{
	RequestSender types.RequestSender
}

func NewDap(requestSender types.RequestSender) Dap {
	return Dap{
		RequestSender: requestSender,
	}
}

func (d Dap) Listen(listener net.Listener) {
	s := rpc.NewServer()
	s.Register(d)
	s.Accept(listener)
}

func (d Dap) Continue(_ struct{}, _ *struct{}) error {
	// TODO: call d.Continue() somehow without an import cycle
	_, err := (d.RequestSender)("continue", nil)
	return err
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

func (d DapClient) Continue() {
	if d.c == nil {
		return
	}
	if err := d.c.Call("Dap.Continue", struct{}{}, nil); err != nil {
		log.Printf("Dap.Continue: %s", err)
	}
}
