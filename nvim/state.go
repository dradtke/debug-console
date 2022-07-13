package nvim

import (
	"sync"

	"github.com/dradtke/debug-console/dap"
)

var (
	// Current state global
	state   State
	stateMu sync.Mutex

	// Registered response handlers
	responseHandlers = make(map[int64]func(*dap.Process, dap.Response))
)

type State struct {
	Running       bool
	Process       *dap.Process
	Capabilities  map[string]bool
	OnInitialized dap.OnInitializedFunc
	Filepath      string
}
