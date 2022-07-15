package nvim

import (
	"encoding/json"
	"log"

	"github.com/dradtke/debug-console/dap"
	"github.com/neovim/go-client/nvim"
)

func HandleEvent(v *nvim.Nvim, d *dap.DAP) func(dap.Event) {
	return func(event dap.Event) {
		switch event.Event {
		case "output":
			var body struct {
				Category string `json:"category"`
				Output   string `json:"output"`
			}
			if err := json.Unmarshal(event.Body, &body); err != nil {
				log.Printf("Error parsing event output: %s", err)
			} else {
				// TODO: check if category is stdout or stderr
				log.Print(body.Output)
			}

		case "terminated":
			log.Print("Debug adapter terminated.")
			d.ClearProcess()

		case "initialized":
			log.Print("Debug adapter initialized")

		case "stopped":
			log.Print("Stopped")
			var body struct {
				AllThreadsStopped bool   `json:"allThreadsStopped"`
				Reason            string `json:"reason"`
				ThreadID          int    `json:"threadId"`
			}
			if err := json.Unmarshal(event.Body, &body); err != nil {
				log.Printf("Error parsing body: %s", err)
			}
			handleStopped(v, body.Reason)

		default:
			log.Printf("Don't know how to handle event: %s", event.Event)
		}
	}
}

func handleStopped(v *nvim.Nvim, reason string) {
	switch reason {
	case "breakpoint":
		log.Print("Stopped at a breakpoint")
		v.Notify("Stopped at a breakpoint", nvim.LogInfoLevel, make(map[string]any))
		// TODO: get breakpoint information
	}
}

// TODO: add a request handler for requests coming from the debug adapter
