package util

import (
	"log"
	"runtime/debug"
)

func LogPanic() {
	if r := recover(); r != nil {
		log.Printf("recovered panic: %v", r)
		log.Print(string(debug.Stack()))
	}
}
