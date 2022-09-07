package nvim

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dradtke/debug-console/dap"
	"github.com/dradtke/debug-console/types"
	"github.com/neovim/go-client/nvim"
)

// TODO: move non-Neovim-specific event handling to the dap package
func HandleEvent(v *nvim.Nvim, d *dap.DAP) types.EventHandler {
	return func(event types.Event) {
		log.Printf("received event: %s", event.Event)
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
			RemoveAllSigns(v, SignGroupCurrentLocation)
			log.Print("Debug adapter terminated.")
			d.Stop()

		case "initialized":
			log.Print("Debug adapter initialized")

		case "stopped":
			var stopped types.StoppedEvent
			if err := json.Unmarshal(event.Body, &stopped); err != nil {
				log.Printf("Error parsing body: %s", err)
			}
			go func() {
				stackFrame, err := d.HandleStopped(stopped)
				if err != nil {
					log.Printf("Error handling stop: %s", err)
					return
				}
				if stackFrame.Source.Name != nil {
					msg := fmt.Sprintf("Stopped (%s) at %s:%d", stopped.Reason, *stackFrame.Source.Name, stackFrame.Line)
					Notify(v, msg, nvim.LogInfoLevel)
				}
				if stackFrame.Source.Path != nil {
					RemoveAllSigns(v, SignGroupCurrentLocation)
					if err := PlaceSign(v, SignNameCurrentLocation, SignInfo{
						Group:         SignGroupCurrentLocation,
						BufferPattern: *stackFrame.Source.Path,
						LineNumber:    stackFrame.Line,
					}, 99); err != nil {
						log.Printf("Error placing current location sign: %s", err)
					}
				}
			}()

		case "continued":
			// TODO: remove current location?
			RemoveAllSigns(v, SignGroupCurrentLocation)

		default:
			log.Printf("Don't know how to handle event: %s", event.Event)
		}
	}
}

// TODO: add a request handler for requests coming from the debug adapter
