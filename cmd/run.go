package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared/errfmt"
)

// Run executes the CLI using the provided args (not including argv[0]) and version string.
// It returns the intended process exit code.
func Run(args []string, versionInfo string) int {
	root := RootCommand(versionInfo)
	defer CleanupTempPrivateKeys()

	if err := root.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
		return 1
	}

	if err := root.Run(context.Background()); err != nil {
		var reported ReportedError
		if errors.As(err, &reported) {
			return 1
		}
		if errors.Is(err, flag.ErrHelp) {
			return 1
		}
		fmt.Fprint(os.Stderr, errfmt.FormatStderr(err))
		return 1
	}

	return 0
}

