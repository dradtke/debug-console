package nvim

import (
	"encoding/json"
	"log"

	"github.com/dradtke/debug-console/dap"
)

func HandleEvent(event dap.Event) {
	switch event.Event {
	case "output":
		var outputBody struct {
			Category string `json:"category"`
			Output   string `json:"output"`
		}
		if err := json.Unmarshal(event.Body, &outputBody); err != nil {
			log.Printf("Error parsing event output: %s", err)
		} else {
			// TODO: check if category is stdout or stderr
			log.Print(outputBody.Output)
		}

	case "terminated":
		log.Print("Debug adapter terminated.")
		dap.ClearState()

	case "initialized":
		log.Print("Debug adapter initialized")

	default:
		log.Printf("Don't know how to handle event: %s", event.Event)
	}
}
