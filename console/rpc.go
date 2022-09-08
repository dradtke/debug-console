package console

import (
	"net/rpc"
	"os"
	"time"

	"github.com/chzyer/readline"
)

type ConsoleService struct {
	Prompt *readline.Instance
	Stops  chan struct{}
}

func NewConsole(dapClient *rpc.Client) (ConsoleService, error) {
	rl, err := readline.New("debug> ")
	if err != nil {
		return ConsoleService{}, err
	}
	rl.Config.AutoComplete = completer{dapClient}
	return ConsoleService{
		Prompt: rl,
		Stops:  make(chan struct{}, 1),
	}, nil
}

func (c ConsoleService) Stop(_ struct{}, _ *struct{}) error {
	// Allow the function to return before exiting in order to avoid an "unexpected EOF" error.
	go func() {
		time.Sleep(100 * time.Millisecond)
		c.Prompt.Close()
		os.Exit(0)
	}()
	return nil
}

func (c ConsoleService) HandleStopped(_ struct{}, _ *struct{}) error {
	c.Stops <- struct{}{}
	return nil
}
