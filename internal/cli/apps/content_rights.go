package apps

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

// AppsContentRightsCommand returns the content-rights command group.
func AppsContentRightsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("content-rights", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "content-rights",
		ShortUsage: "asc apps content-rights <subcommand> [flags]",
		ShortHelp:  "Manage an app's content rights declaration.",
		LongHelp: `Manage an app's content rights declaration.

The content rights declaration indicates whether your app uses third-party content.
This is required for App Store submission.

Examples:
  asc apps content-rights get --app "APP_ID"
  asc apps content-rights set --app "APP_ID" --uses-third-party-content=false`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			AppsContentRightsGetCommand(),
			AppsContentRightsSetCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// AppsContentRightsGetCommand returns the content-rights get subcommand.
func AppsContentRightsGetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("content-rights get", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "get",
		ShortUsage: "asc apps content-rights get --app \"APP_ID\"",
		ShortHelp:  "Get an app's content rights declaration.",
		LongHelp: `Get an app's content rights declaration.

Examples:
  asc apps content-rights get --app "APP_ID"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := shared.ResolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("apps content-rights get: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			app, err := client.GetApp(requestCtx, resolvedAppID)
			if err != nil {
				return fmt.Errorf("apps content-rights get: failed to fetch: %w", err)
			}

			return shared.PrintOutput(app, *output.Output, *output.Pretty)
		},
	}
}

// AppsContentRightsSetCommand returns the content-rights set subcommand.
func AppsContentRightsSetCommand() *ffcli.Command {
	fs := flag.NewFlagSet("content-rights set", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	usesThirdParty := fs.String("uses-third-party-content", "", "Whether app uses third-party content (true/false, yes/no)")
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "set",
		ShortUsage: "asc apps content-rights set --app \"APP_ID\" --uses-third-party-content false",
		ShortHelp:  "Set an app's content rights declaration.",
		LongHelp: `Set an app's content rights declaration.

This declares whether your app uses third-party content, which is required
for App Store submission.

  --uses-third-party-content false  → DOES_NOT_USE_THIRD_PARTY_CONTENT
  --uses-third-party-content true   → USES_THIRD_PARTY_CONTENT

Accepts: true, false, yes, no, uses, does-not-use, or the raw API values.

Examples:
  asc apps content-rights set --app "APP_ID" --uses-third-party-content false
  asc apps content-rights set --app "APP_ID" --uses-third-party-content true`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			resolvedAppID := shared.ResolveAppID(*appID)
			if resolvedAppID == "" {
				fmt.Fprintln(os.Stderr, "Error: --app is required (or set ASC_APP_ID)")
				return flag.ErrHelp
			}

			if strings.TrimSpace(*usesThirdParty) == "" {
				fmt.Fprintln(os.Stderr, "Error: --uses-third-party-content is required (true or false)")
				return flag.ErrHelp
			}

			declaration, err := parseContentRightsValue(*usesThirdParty)
			if err != nil {
				return shared.UsageError(err.Error())
			}

			client, err := shared.GetASCClient()
			if err != nil {
				return fmt.Errorf("apps content-rights set: %w", err)
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			attrs := asc.AppUpdateAttributes{
				ContentRightsDeclaration: &declaration,
			}

			app, err := client.UpdateApp(requestCtx, resolvedAppID, attrs)
			if err != nil {
				return fmt.Errorf("apps content-rights set: failed to update: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Content rights declaration set to %s\n", string(declaration))

			return shared.PrintOutput(app, *output.Output, *output.Pretty)
		},
	}
}

// parseContentRightsValue converts a user-friendly string to the API enum.
// Accepts: true, false, yes, no, uses, does-not-use, and the raw API values.
func parseContentRightsValue(s string) (asc.ContentRightsDeclaration, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "false", "no", "does-not-use", "does_not_use_third_party_content":
		return asc.ContentRightsDeclarationDoesNotUseThirdPartyContent, nil
	case "true", "yes", "uses", "uses_third_party_content":
		return asc.ContentRightsDeclarationUsesThirdPartyContent, nil
	default:
		return "", fmt.Errorf("invalid value %q: use true/false, yes/no, or DOES_NOT_USE_THIRD_PARTY_CONTENT/USES_THIRD_PARTY_CONTENT", s)
	}
}
