package dap

import "encoding/gob"

func init() {
	gob.Register(Output{})
}
