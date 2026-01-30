package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/registry"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared/suggest"
)

// RootCommand returns the root command
func RootCommand(version string) *ffcli.Command {
	root := &ffcli.Command{
		Name:        "asc",
		ShortUsage:  "asc <subcommand> [flags]",
		ShortHelp:   "A fast, AI-agent friendly CLI for App Store Connect.",
		LongHelp:    "ASC is a lightweight CLI for App Store Connect. Built for developers and AI agents.",
		FlagSet:     flag.NewFlagSet("asc", flag.ExitOnError),
		UsageFunc:   DefaultUsageFunc,
		Subcommands: registry.Subcommands(version),
	}

	versionFlag := root.FlagSet.Bool("version", false, "Print version and exit")
	shared.BindRootFlags(root.FlagSet)

	rootSubcommandNames := make([]string, 0, len(root.Subcommands))
	for _, sub := range root.Subcommands {
		rootSubcommandNames = append(rootSubcommandNames, sub.Name)
	}

	root.Exec = func(ctx context.Context, args []string) error {
		if *versionFlag {
			fmt.Fprintln(os.Stdout, version)
			return nil
		}
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
			if suggestions := suggest.Commands(args[0], rootSubcommandNames); len(suggestions) > 0 {
				fmt.Fprintf(os.Stderr, "Did you mean: %s\n\n", strings.Join(suggestions, ", "))
			}
		}
		return flag.ErrHelp
	}

	return root
}
