package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/picker"
	"github.com/spf13/cobra"
)

// resolveSnippet resolves a reference to a single snippet using the resolution strategy:
//  1. Single result -> use it
//  2. Single high-confidence match (top score >= 2x second-best) -> use it
//  3. Multiple results + TTY -> launch mini-picker
//  4. Multiple results + non-TTY -> take top match
//  5. Zero results -> exit code 2
func resolveSnippet(cmd *cobra.Command, ref string) (*model.Snippet, error) {
	results, err := store.ResolveRef(ref)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "no snippet matching %q\n", ref)
		os.Exit(model.ExitNotFound)
	}

	// Single result -- use it.
	if len(results) == 1 {
		return &results[0].Snippet, nil
	}

	// High-confidence: top score >= 2x second-best.
	if results[0].Score >= 2*results[1].Score {
		return &results[0].Snippet, nil
	}

	// Multiple ambiguous results.
	// Check if stdout is a TTY.
	stat, _ := os.Stdout.Stat()
	isTTY := stat.Mode()&os.ModeCharDevice != 0

	if isTTY {
		// Launch mini-picker.
		selected, err := picker.RunPicker(results, ref)
		if err != nil {
			return nil, fmt.Errorf("picker: %w", err)
		}
		if selected == nil {
			os.Exit(model.ExitInterrupted)
		}
		return selected, nil
	}

	// Non-TTY: take top match.
	return &results[0].Snippet, nil
}
