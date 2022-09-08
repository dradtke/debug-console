package dap

import (
	"encoding/json"
	"log"

	"github.com/dradtke/debug-console/types"
)

func (d *DAP) handleEvent(event types.Event) {
	switch event.Event {
	case "initialized":
		log.Print("Debug adapter initialized")

	case "terminated":
		log.Print("Debug adapter terminated.")
		d.Stop()

	case "output":
		var body types.OutputEvent
		if err := json.Unmarshal(event.Body, &body); err != nil {
			log.Printf("Error parsing event output: %s", err)
		} else {
			if err := d.ShowOutput(body); err != nil {
				log.Printf("Error showing event output: %s", err)
			}
		}
	}
}
