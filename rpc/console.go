package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
)

type Console struct{}

func (c Console) ShowOutput(output string, _ *struct{}) error {
	log.Printf("output: %s", output)
	return nil
}

func (c Console) Quit(_ struct{}, _ *struct{}) error {
	log.Print("Quitting")
	// Allow the function to return before exiting in order to avoid an "unexpected EOF" error.
	go func() {
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
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

func (c ConsoleClient) Quit() {
	if err := c.c.Call("Console.Quit", struct{}{}, nil); err != nil {
		log.Printf("Error quitting debug console: %s", err)
	}
}
