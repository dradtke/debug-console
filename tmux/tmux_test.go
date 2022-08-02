package tmux_test

import (
	"testing"

	"github.com/dradtke/debug-console/tmux"
)

func TestSendKeys(t *testing.T) {
	tmux.RunInPane("1", "echo", "hello world")
}
