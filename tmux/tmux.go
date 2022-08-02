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
		{"tmux", "select-pane", "-T", "console"},
		{"tmux", "select-pane", "-l"},
	})
}

// FindPane returns the id of the pane with the given title.
func FindPane(title string) (string, error) {
	v, err := exec.Command("tmux", "list-panes", "-f", fmt.Sprintf("#{==:#{pane_title},%s}", title), "-F", "#{pane_id}").Output()
	return strings.TrimSpace(string(v)), err
}

func RunInPane(pane string, args ...string) error {
	tmuxArgs := []string{"send-keys", "-t", pane}
	for i := range args {
		// Not sure if this adheres to shell rules, but it seems to work okay.
		args[i] = strconv.Quote(args[i])
	}
	tmuxArgs = append(tmuxArgs, strings.Join(args, " "))
	tmuxArgs = append(tmuxArgs, "Enter")
	return exec.Command("tmux", tmuxArgs...).Run()
}

func runAll(cmds [][]string) error {
	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("Error executing command: %s: %w", strings.Join(cmd, " "), err)
		}
	}
	return nil
}
