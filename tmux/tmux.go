package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func NumPanes() (int, error) {
	b, err := exec.Command("tmux", "display-message", "-p", "#{window_panes}").Output()
	if err != nil {
		return 0, fmt.Errorf("Error running command: %w", err)
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, fmt.Errorf("Error parsing result: %s", string(b))
	}
	return v, nil
}

func Split() error {
	return runAll([][]string{
		{"tmux", "split-pane", "-p", "40", "-h"},
		{"tmux", "select-pane", "-t", "0"},
	})
}

func runAll(cmds [][]string) error {
	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("Error executing command: %s: %w", strings.Join(cmd, " "), err)
		}
	}
	return nil
}
