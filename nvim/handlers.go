package nvim

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/dradtke/debug-console/dap"
)

func HandleResponse(p *dap.Process, resp dap.Response) {
	if !resp.Success {
		handleResponseError(resp.Body)
		return
	}

	switch resp.Command {
	case "initialize":
		log.Print("Initialized successfully!")
		state.Capabilities = make(map[string]bool)
		if err := json.Unmarshal(resp.Body, &state.Capabilities); err != nil {
			log.Printf("Error parsing capabilities: %s", err)
		}
		if state.OnInitialized != nil {
			state.OnInitialized(state.Filepath, state.Process)
		}

	case "launch":
		log.Print("Debug adapter launched!")
		// TODO: set all breakpoints
		// For now, just send configurationDone
		p.SendRequest("configurationDone", make(map[string]any))

	default:
		log.Printf("Don't know how to handle response: %s", resp.Command)
	}
}

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
		state = State{}

	case "initialized":
		log.Print("Debug adapter initialized")

	default:
		log.Printf("Don't know how to handle event: %s", event.Event)
	}
}

func handleResponseError(body json.RawMessage) {
	var errBody struct {
		Error struct {
			ID        int               `json:"id"`
			Format    string            `json:"format"`
			Variables map[string]string `json:"variables"`
		} `json:"error"`
		ShowUser bool `json:"showUser"`
	}
	if err := json.Unmarshal(body, &errBody); err != nil {
		log.Printf("Error parsing error body: %s", err)
	}
	msg := errBody.Error.Format
	for name, value := range errBody.Error.Variables {
		msg = strings.ReplaceAll(msg, "{"+name+"}", value)
	}
	// TODO: if errBody.ShowUser, send a notification to the UI
	log.Printf("%d: %s", errBody.Error.ID, msg)
}
