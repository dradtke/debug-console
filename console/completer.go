package console

import (
	"log"
	"net/rpc"

	"github.com/dradtke/debug-console/types"
)

type completer struct {
	dapClient *rpc.Client
}

func (c completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	var (
		items []types.CompletionItem
		args  = types.CompletionsArguments{
			Text:   string(line),
			Column: len(line)-1,
		}
	)
	if err := c.dapClient.Call("DAPService.Completions", args, &items); err != nil {
		log.Println(err)
		return nil, 0
	}

	for _, item := range items {
		newLine = append(newLine, []rune(item.Label))
	}
	return newLine, len(line)
}
