package util

import (
	"log"
	"runtime/debug"
)

func Recover() {
	if r := recover(); r != nil {
		LogPanic(r)
	}
}

func LogPanic(v any) {
	log.Printf("panic: %v", v)
	log.Print(string(debug.Stack()))
}
