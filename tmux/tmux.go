package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ShellEscapeFunc defines the function to use for escaping shell commands. It
// uses strconv.Quote() by default, but can be overridden with an
// editor-specific function.
var ShellEscapeFunc = strconv.Quote

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

// TODO: ensure that these splits only take effect in the same tmux pane as the editor,
// even if it is not currently selected.
func SplitConsole() error {
	return runAll([][]string{
		{"tmux", "split-pane", "-p", "40", "-h"},
		{"tmux", "select-pane", "-T", "console"},
		{"tmux", "select-pane", "-l"},
	})
}

func SplitOutput() error {
	return runAll([][]string{
		{"tmux", "select-pane", "-t", "right"},
		{"tmux", "split-pane", "-p", "40", "-v"},
		{"tmux", "select-pane", "-T", "output"},
		{"tmux", "select-pane", "-t", "left"},
	})
}

func SplitRunInTerminal() error {
	return runAll([][]string{
		{"tmux", "select-pane", "-t", "right"},
		{"tmux", "split-pane", "-p", "40", "-v"},
		{"tmux", "select-pane", "-T", "run-in-terminal"},
		{"tmux", "select-pane", "-t", "left"},
	})
}

func FindOrSplitOutput() (string, error) {
	if pane, err := FindPane("output"); err != nil {
		return "", err
	} else if pane != "" {
		return pane, nil
	}

	if err := SplitOutput(); err != nil {
		return "", err
	}
	return FindPane("output")
}

func FindOrSplitRunInTerminal() (string, error) {
	if pane, err := FindPane("run-in-terminal"); err != nil {
		return "", err
	} else if pane != "" {
		return pane, nil
	}

	if err := SplitRunInTerminal(); err != nil {
		return "", err
	}
	return FindPane("run-in-terminal")
}

// FindPane returns the id of the pane with the given title.
func FindPane(title string) (string, error) {
	v, err := exec.Command("tmux", "list-panes", "-f", fmt.Sprintf("#{==:#{pane_title},%s}", title), "-F", "#{pane_id}").Output()
	return strings.TrimSpace(string(v)), err
}

func RunInPane(pane string, args ...string) error {
	tmuxArgs := []string{"send-keys", "-t", pane}
	for i := range args {
		args[i] = ShellEscapeFunc(args[i])
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
