package nvim

import (
	"encoding/json"
	"log"

	"github.com/dradtke/debug-console/dap"
	"github.com/dradtke/debug-console/types"
	"github.com/neovim/go-client/nvim"
)

// TODO: move non-Neovim-specific event handling to the dap package
func HandleEvent(v *nvim.Nvim, d *dap.DAP) types.EventHandler {
	return func(event types.Event) {
		switch event.Event {
		case "output":
			var body dap.Output
			if err := json.Unmarshal(event.Body, &body); err != nil {
				log.Printf("Error parsing event output: %s", err)
			} else {
				if err := d.ShowOutput(body); err != nil {
					log.Printf("Error showing event output: %s", err)
				}
			}

		case "terminated":
			log.Print("Debug adapter terminated.")
			d.ConsoleClient.Stop()
			d.OutputBroadcaster.Stop()
			// ???: Is this the correct behavior?
			d.Conn.Stop()

		case "initialized":
			log.Print("Debug adapter initialized")

		case "stopped":
			log.Print("Stopped")
			var body dap.Stopped
			if err := json.Unmarshal(event.Body, &body); err != nil {
				log.Printf("Error parsing body: %s", err)
			}
			handleStopped(v, d, body.Reason)

		default:
			log.Printf("Don't know how to handle event: %s", event.Event)
		}
	}
}

func handleStopped(v *nvim.Nvim, d *dap.DAP, reason string) {
	switch reason {
	case "breakpoint":
		v.Notify("Stopped at a breakpoint", nvim.LogInfoLevel, make(map[string]any))
		d.ConsoleClient.HandleStopped()
		// TODO: get breakpoint information

	default:
		log.Printf("Stopped for unknown reason: %s", reason)
	}
}

// TODO: add a request handler for requests coming from the debug adapter
