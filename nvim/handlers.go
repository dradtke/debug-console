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
			log.Printf("%+v", body)
			handleStopped(v, d, body)

		default:
			log.Printf("Don't know how to handle event: %s", event.Event)
		}
	}
}

func handleStopped(v *nvim.Nvim, d *dap.DAP, body dap.Stopped) {
	d.ConsoleClient.HandleStopped()
	if body.ThreadID != nil {
		go getStoppedAt(v, d, body)
	}
}

func getStoppedAt(v *nvim.Nvim, d *dap.DAP, stopped dap.Stopped) {
	resp, err := d.SendRequest("stackTrace", map[string]any{
		"threadId": *stopped.ThreadID,
		"levels":   1,
		"format": map[string]any{
			"line": true,
		},
	})
	if err != nil {
		log.Printf("Error getting stack trace: %s", err)
		return
	}

	var body struct {
		StackFrames []types.StackFrame `json:"stackFrames"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		log.Printf("Error parsing stackTrace response: %s", err)
		return
	}

	if len(body.StackFrames) > 0 {
		stackFrame := body.StackFrames[0]
		d.Lock()
		d.StoppedLocation = &stackFrame
		d.Unlock()
		if stackFrame.Source.Name != nil {
			msg := fmt.Sprintf("Stopped (%s) at %s:%d", stopped.Reason, *stackFrame.Source.Name, stackFrame.Line)
			Notify(v, msg, nvim.LogInfoLevel)
		}
	}
}

// TODO: add a request handler for requests coming from the debug adapter
