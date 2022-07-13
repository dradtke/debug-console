package nvim

import (
	"github.com/dradtke/debug-console/dap"
)

// Current state global
var state State

type State struct {
	Running       bool
	Process       *dap.Process
	Capabilities  map[string]bool
	OnInitialized dap.OnInitializedFunc
	Filepath      string
}
