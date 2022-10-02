package dap

import (
	"encoding/json"
	"log"

	"github.com/dradtke/debug-console/types"
)

func (d *DAP) HandleEvent(event types.Event) {
	log.Printf("Received event: %s", event.Event)

	switch event.Event {
	case "initialized":
		log.Print("Debug adapter initialized")
		d.Lock()
		defer d.Unlock()
		if d.Conn != nil {
			d.Conn.seeInitializeEvent.Do(func() {
				log.Println("Closing the channel")
				close(d.Conn.initializedEventSeen)
			})
		}

	case "output":
		var output types.OutputEvent
		if err := json.Unmarshal(event.Body, &output); err != nil {
			log.Printf("Error parsing output event: %s", err)
		} else if err = d.ShowOutput(output); err != nil {
			log.Printf("Error showing output: %s", err)
		}

	case "terminated":
		log.Print("Debug adapter terminated")
		d.Stop()
	}
}
