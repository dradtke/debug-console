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
		switch event.Event {
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
			RemoveAllSigns(v, SignGroupCurrentLocation)

		case "terminated":
			Notify(v, "Debug adapter terminated", nvim.LogInfoLevel)
			RemoveAllSigns(v, SignGroupCurrentLocation)
		}
	}
}

// TODO: add a request handler for requests coming from the debug adapter
