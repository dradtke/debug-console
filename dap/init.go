package dap

import (
	"encoding/gob"

	"github.com/dradtke/debug-console/types"
)

func init() {
	gob.Register(types.OutputEvent{})
}
