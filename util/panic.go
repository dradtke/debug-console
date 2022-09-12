package util

import "log"

func LogPanic() {
	if r := recover(); r != nil {
		log.Printf("panic: %v", r)
	}
}
