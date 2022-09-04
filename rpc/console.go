package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/chzyer/readline"
)

type Console struct{
	Prompt *readline.Instance
	Stops chan struct{}
}

func NewConsole() (Console, error) {
	rl, err := readline.New("> ")
	if err != nil {
		return Console{}, err
	}
	return Console{
		Prompt: rl,
		Stops: make(chan struct{}, 1),
	}, nil
}

func (c Console) Stop(_ struct{}, _ *struct{}) error {
	// Allow the function to return before exiting in order to avoid an "unexpected EOF" error.
	go func() {
		time.Sleep(100 * time.Millisecond)
		c.Prompt.Close()
		os.Exit(0)
	}()
	return nil
}

func (c Console) HandleStopped(_ struct{}, _ *struct{}) error {
	c.Stops <- struct{}{}
	return nil
}

func (c Console) Listen(listener net.Listener) {
	s := rpc.NewServer()
	s.Register(c)
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

func (c ConsoleClient) Stop() {
	if c.c == nil {
		return
	}
	if err := c.c.Call("Console.Stop", struct{}{}, nil); err != nil {
		log.Printf("Error quitting debug console: %s", err)
	}
}

func (c ConsoleClient) HandleStopped() {
	if c.c == nil {
		return
	}
	if err := c.c.Call("Console.HandleStopped", struct{}{}, nil); err != nil {
		log.Printf("Console.HandleStopped: %s", err)
	}
}
