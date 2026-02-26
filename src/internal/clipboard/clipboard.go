package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type tool struct {
	copyCmd   string
	copyArgs  []string
	pasteCmd  string
	pasteArgs []string
}

func detect() *tool {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("pbcopy"); err == nil {
			return &tool{
				copyCmd: "pbcopy", copyArgs: nil,
				pasteCmd: "pbpaste", pasteArgs: nil,
			}
		}
	default: // linux, etc.
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return &tool{
				copyCmd: "wl-copy", copyArgs: nil,
				pasteCmd: "wl-paste", pasteArgs: []string{"--no-newline"},
			}
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return &tool{
				copyCmd: "xclip", copyArgs: []string{"-selection", "clipboard"},
				pasteCmd: "xclip", pasteArgs: []string{"-selection", "clipboard", "-o"},
			}
		}
	}
	return nil
}

// Available returns true if a clipboard tool is detected.
func Available() bool {
	return detect() != nil
}

// Write copies text to the system clipboard.
func Write(text string) error {
	t := detect()
	if t == nil {
		return fmt.Errorf("no clipboard tool found (need pbcopy, xclip, or wl-copy)")
	}
	cmd := exec.Command(t.copyCmd, t.copyArgs...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// Read returns the current clipboard contents.
func Read() (string, error) {
	t := detect()
	if t == nil {
		return "", fmt.Errorf("no clipboard tool found (need pbcopy, xclip, or wl-copy)")
	}
	cmd := exec.Command(t.pasteCmd, t.pasteArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("clipboard read failed: %w", err)
	}
	return string(out), nil
}
