//go:build !nvim

package main

import "errors"

func runNvim(args []string) error {
	return errors.New("built without nvim feature!")
}
