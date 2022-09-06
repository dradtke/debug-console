package nvim

import (
	"log"

	"github.com/neovim/go-client/nvim"
)

func Notify(v *nvim.Nvim, msg string, logLevel nvim.LogLevel) {
	if err := v.Notify(msg, logLevel, make(map[string]any)); err != nil {
		log.Printf("Failed to send notification: %s", err)
	}
}
