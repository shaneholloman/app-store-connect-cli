package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/auth"
)

// Feedback command factory
func FeedbackCommand() *ffcli.Command {
	fs := flag.NewFlagSet("feedback", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	deviceModel := fs.String("device-model", "", "Filter by device model(s), comma-separated")
	osVersion := fs.String("os-version", "", "Filter by OS version(s), comma-separated")
	appPlatform := fs.String("app-platform", "", "Filter by app platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	devicePlatform := fs.String("device-platform", "", "Filter by device platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	buildID := fs.String("build", "", "Filter by build ID(s), comma-separated")
	buildPreRelease := fs.String("build-pre-release-version", "", "Filter by pre-release version ID(s), comma-separated")
	tester := fs.String("tester", "", "Filter by tester ID(s), comma-separated")
	sort := fs.String("sort", "", "Sort by createdDate or -createdDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "feedback",
		ShortUsage: "asc feedback [flags]",
		ShortHelp:  "List TestFlight feedback from beta testers.",
		LongHelp: `List TestFlight feedback from beta testers.

This command fetches beta feedback screenshot submissions and comments.

Examples:
  asc feedback --app "123456789"
  asc feedback --app "123456789" --json
  asc feedback --app "123456789" --device-model "iPhone15,3" --os-version "17.2"
  asc feedback --app "123456789" --sort -createdDate --limit 5 --json
  asc feedback --next "<links.next>" --json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("feedback: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("feedback: %w", err)
			}
			if err := validateSort(*sort, "createdDate", "-createdDate"); err != nil {
				return fmt.Errorf("feedback: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("feedback: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.FeedbackOption{
				asc.WithFeedbackDeviceModels(splitCSV(*deviceModel)),
				asc.WithFeedbackOSVersions(splitCSV(*osVersion)),
				asc.WithFeedbackAppPlatforms(splitCSVUpper(*appPlatform)),
				asc.WithFeedbackDevicePlatforms(splitCSVUpper(*devicePlatform)),
				asc.WithFeedbackBuildIDs(splitCSV(*buildID)),
				asc.WithFeedbackBuildPreReleaseVersionIDs(splitCSV(*buildPreRelease)),
				asc.WithFeedbackTesterIDs(splitCSV(*tester)),
				asc.WithFeedbackLimit(*limit),
				asc.WithFeedbackNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithFeedbackSort(*sort))
			}

			feedback, err := client.GetFeedback(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("feedback: failed to fetch: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(feedback, format, *pretty)
		},
	}
}

// Crashes command factory
func CrashesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("crashes", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	deviceModel := fs.String("device-model", "", "Filter by device model(s), comma-separated")
	osVersion := fs.String("os-version", "", "Filter by OS version(s), comma-separated")
	appPlatform := fs.String("app-platform", "", "Filter by app platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	devicePlatform := fs.String("device-platform", "", "Filter by device platform(s), comma-separated (IOS, MAC_OS, TV_OS, VISION_OS)")
	buildID := fs.String("build", "", "Filter by build ID(s), comma-separated")
	buildPreRelease := fs.String("build-pre-release-version", "", "Filter by pre-release version ID(s), comma-separated")
	tester := fs.String("tester", "", "Filter by tester ID(s), comma-separated")
	sort := fs.String("sort", "", "Sort by createdDate or -createdDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "crashes",
		ShortUsage: "asc crashes [flags]",
		ShortHelp:  "List and export TestFlight crash reports.",
		LongHelp: `List and export TestFlight crash reports.

This command fetches crash reports submitted by TestFlight beta testers,
helping you identify and fix issues in your app.

Examples:
  asc crashes --app "123456789"
  asc crashes --app "123456789" --json
  asc crashes --app "123456789" --json > crashes.json
  asc crashes --app "123456789" --device-model "iPhone15,3" --os-version "17.2"
  asc crashes --app "123456789" --sort -createdDate --limit 5 --json
  asc crashes --next "<links.next>" --json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("crashes: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("crashes: %w", err)
			}
			if err := validateSort(*sort, "createdDate", "-createdDate"); err != nil {
				return fmt.Errorf("crashes: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("crashes: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.CrashOption{
				asc.WithCrashDeviceModels(splitCSV(*deviceModel)),
				asc.WithCrashOSVersions(splitCSV(*osVersion)),
				asc.WithCrashAppPlatforms(splitCSVUpper(*appPlatform)),
				asc.WithCrashDevicePlatforms(splitCSVUpper(*devicePlatform)),
				asc.WithCrashBuildIDs(splitCSV(*buildID)),
				asc.WithCrashBuildPreReleaseVersionIDs(splitCSV(*buildPreRelease)),
				asc.WithCrashTesterIDs(splitCSV(*tester)),
				asc.WithCrashLimit(*limit),
				asc.WithCrashNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithCrashSort(*sort))
			}

			crashes, err := client.GetCrashes(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("crashes: failed to fetch: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(crashes, format, *pretty)
		},
	}
}

// Reviews command factory
func ReviewsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("reviews", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	stars := fs.Int("stars", 0, "Filter by star rating (1-5)")
	territory := fs.String("territory", "", "Filter by territory (e.g., US, GBR)")
	sort := fs.String("sort", "", "Sort by rating, -rating, createdDate, or -createdDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "reviews",
		ShortUsage: "asc reviews [flags]",
		ShortHelp:  "List App Store customer reviews.",
		LongHelp: `List App Store customer reviews.

This command fetches customer reviews from the App Store,
helping you understand user feedback and sentiment.

Examples:
  asc reviews --app "123456789"
  asc reviews --app "123456789" --json
  asc reviews --app "123456789" --stars 1 --territory US --json
  asc reviews --app "123456789" --sort -createdDate --limit 5 --json
  asc reviews --next "<links.next>" --json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("reviews: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("reviews: %w", err)
			}
			if err := validateSort(*sort, "rating", "-rating", "createdDate", "-createdDate"); err != nil {
				return fmt.Errorf("reviews: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("reviews: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.ReviewOption{}
			if *stars != 0 {
				if *stars < 1 || *stars > 5 {
					return fmt.Errorf("reviews: --stars must be between 1 and 5")
				}
				opts = append(opts, asc.WithRating(*stars))
			}
			if *territory != "" {
				opts = append(opts, asc.WithTerritory(*territory))
			}
			if *limit != 0 {
				opts = append(opts, asc.WithLimit(*limit))
			}
			if strings.TrimSpace(*next) != "" {
				opts = append(opts, asc.WithNextURL(*next))
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithReviewSort(*sort))
			}

			reviews, err := client.GetReviews(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("reviews: failed to fetch: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(reviews, format, *pretty)
		},
	}
}

// Apps command factory
func AppsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("apps", flag.ExitOnError)

	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	sort := fs.String("sort", "", "Sort by name or -name, bundleId or -bundleId")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "apps",
		ShortUsage: "asc apps [flags]",
		ShortHelp:  "List all apps in your App Store Connect account.",
		LongHelp: `List all apps in your App Store Connect account.

This command fetches all apps associated with your API key,
useful for finding app IDs when using other commands.

Examples:
  asc apps
  asc apps --json
  asc apps --limit 10 --json
  asc apps --sort name --json
  asc apps --output table
  asc apps --next "<links.next>" --json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("apps: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("apps: %w", err)
			}
			if err := validateSort(*sort, "name", "-name", "bundleId", "-bundleId"); err != nil {
				return fmt.Errorf("apps: %w", err)
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("apps: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.AppsOption{
				asc.WithAppsLimit(*limit),
				asc.WithAppsNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithAppsSort(*sort))
			}

			apps, err := client.GetApps(requestCtx, opts...)
			if err != nil {
				return fmt.Errorf("apps: failed to fetch: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(apps, format, *pretty)
		},
	}
}

// Builds command factory
func BuildsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("builds", flag.ExitOnError)

	appID := fs.String("app", "", "App Store Connect app ID (or ASC_APP_ID env)")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")
	sort := fs.String("sort", "", "Sort by uploadedDate or -uploadedDate")
	limit := fs.Int("limit", 0, "Maximum results per page (1-200)")
	next := fs.String("next", "", "Fetch next page using a links.next URL")

	return &ffcli.Command{
		Name:       "builds",
		ShortUsage: "asc builds [flags]",
		ShortHelp:  "List builds for an app in App Store Connect.",
		LongHelp: `List builds for an app in App Store Connect.

This command fetches builds uploaded to App Store Connect,
including processing status and expiration dates.

Examples:
  asc builds --app "123456789"
  asc builds --app "123456789" --json
  asc builds --app "123456789" --limit 10 --json
  asc builds --app "123456789" --sort -uploadedDate --json
  asc builds --app "123456789" --output table
  asc builds --next "<links.next>" --json

Subcommands:
  info    Show build details
  expire  Expire a build for TestFlight`,
		FlagSet: fs,
		Subcommands: []*ffcli.Command{
			BuildsInfoCommand(),
			BuildsExpireCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if *limit != 0 && (*limit < 1 || *limit > 200) {
				return fmt.Errorf("builds: --limit must be between 1 and 200")
			}
			if err := validateNextURL(*next); err != nil {
				return fmt.Errorf("builds: %w", err)
			}
			if err := validateSort(*sort, "uploadedDate", "-uploadedDate"); err != nil {
				return fmt.Errorf("builds: %w", err)
			}

			resolvedAppID := resolveAppID(*appID)
			if resolvedAppID == "" && strings.TrimSpace(*next) == "" {
				fmt.Fprintf(os.Stderr, "Error: --app is required (or set ASC_APP_ID)\n\n")
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			opts := []asc.BuildsOption{
				asc.WithBuildsLimit(*limit),
				asc.WithBuildsNextURL(*next),
			}
			if strings.TrimSpace(*sort) != "" {
				opts = append(opts, asc.WithBuildsSort(*sort))
			}

			builds, err := client.GetBuilds(requestCtx, resolvedAppID, opts...)
			if err != nil {
				return fmt.Errorf("builds: failed to fetch: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(builds, format, *pretty)
		},
	}
}

// BuildsInfoCommand returns a build detail subcommand.
func BuildsInfoCommand() *ffcli.Command {
	fs := flag.NewFlagSet("builds info", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "info",
		ShortUsage: "asc builds info [flags]",
		ShortHelp:  "Show build details.",
		LongHelp: `Show build details.

Examples:
  asc builds info --build "BUILD_ID" --json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*buildID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				fs.Usage()
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds info: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			build, err := client.GetBuild(requestCtx, strings.TrimSpace(*buildID))
			if err != nil {
				return fmt.Errorf("builds info: failed to fetch: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(build, format, *pretty)
		},
	}
}

// BuildsExpireCommand returns a build expiration subcommand.
func BuildsExpireCommand() *ffcli.Command {
	fs := flag.NewFlagSet("builds expire", flag.ExitOnError)

	buildID := fs.String("build", "", "Build ID")
	output := fs.String("output", "json", "Output format: json (default), table, markdown")
	jsonFlag := fs.Bool("json", false, "Output in JSON format (shorthand)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	return &ffcli.Command{
		Name:       "expire",
		ShortUsage: "asc builds expire [flags]",
		ShortHelp:  "Expire a build for TestFlight.",
		LongHelp: `Expire a build for TestFlight.

This action is irreversible for the specified build.

Examples:
  asc builds expire --build "BUILD_ID" --json`,
		FlagSet: fs,
		Exec: func(ctx context.Context, args []string) error {
			if strings.TrimSpace(*buildID) == "" {
				fmt.Fprintln(os.Stderr, "Error: --build is required")
				fs.Usage()
				return flag.ErrHelp
			}

			client, err := getASCClient()
			if err != nil {
				return fmt.Errorf("builds expire: %w", err)
			}

			requestCtx, cancel := contextWithTimeout(ctx)
			defer cancel()

			build, err := client.ExpireBuild(requestCtx, strings.TrimSpace(*buildID))
			if err != nil {
				return fmt.Errorf("builds expire: failed to expire: %w", err)
			}

			format := *output
			if *jsonFlag {
				format = "json"
			}

			return printOutput(build, format, *pretty)
		},
	}
}

// VersionCommand returns a version subcommand
func VersionCommand(version string) *ffcli.Command {
	return &ffcli.Command{
		Name:       "version",
		ShortUsage: "asc version",
		ShortHelp:  "Print version information and exit.",
		Exec: func(ctx context.Context, args []string) error {
			fmt.Println(version)
			return nil
		},
	}
}

// RootCommand returns the root command
func RootCommand(version string) *ffcli.Command {
	root := &ffcli.Command{
		Name:       "asc",
		ShortUsage: "asc <subcommand> [flags]",
		ShortHelp:  "A fast, AI-agent friendly CLI for App Store Connect.",
		LongHelp:   "ASC is a lightweight CLI for App Store Connect. Built for developers and AI agents.",
		FlagSet:    flag.NewFlagSet("asc", flag.ExitOnError),
		Subcommands: []*ffcli.Command{
			AuthCommand(),
			FeedbackCommand(),
			CrashesCommand(),
			ReviewsCommand(),
			AppsCommand(),
			BuildsCommand(),
			VersionCommand(version),
		},
	}

	versionFlag := root.FlagSet.Bool("version", false, "Print version and exit")

	root.Exec = func(ctx context.Context, args []string) error {
		if *versionFlag {
			fmt.Fprintln(os.Stdout, version)
			return nil
		}
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		}
		return flag.ErrHelp
	}

	return root
}

func getASCClient() (*asc.Client, error) {
	actualKeyID := os.Getenv("ASC_KEY_ID")
	actualIssuerID := os.Getenv("ASC_ISSUER_ID")
	actualKeyPath := os.Getenv("ASC_PRIVATE_KEY_PATH")

	if actualKeyID == "" || actualIssuerID == "" || actualKeyPath == "" {
		cfg, err := auth.GetDefaultCredentials()
		if err != nil && actualKeyID == "" && actualIssuerID == "" && actualKeyPath == "" {
			return nil, err
		}
		if cfg != nil {
			if actualKeyID == "" {
				actualKeyID = cfg.KeyID
			}
			if actualIssuerID == "" {
				actualIssuerID = cfg.IssuerID
			}
			if actualKeyPath == "" {
				actualKeyPath = cfg.PrivateKeyPath
			}
		}
	}

	if actualKeyID == "" || actualIssuerID == "" || actualKeyPath == "" {
		return nil, fmt.Errorf("missing authentication. Run 'asc auth login'")
	}

	return asc.NewClient(actualKeyID, actualIssuerID, actualKeyPath)
}

func printOutput(data interface{}, format string, pretty bool) error {
	format = strings.ToLower(format)
	switch format {
	case "json":
		if pretty {
			return asc.PrintPrettyJSON(data)
		}
		return asc.PrintJSON(data)
	case "markdown", "md":
		if pretty {
			return fmt.Errorf("--pretty is only valid with JSON output")
		}
		return asc.PrintMarkdown(data)
	case "table":
		if pretty {
			return fmt.Errorf("--pretty is only valid with JSON output")
		}
		return asc.PrintTable(data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func resolveAppID(appID string) string {
	if appID != "" {
		return appID
	}
	return os.Getenv("ASC_APP_ID")
}

func contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, asc.DefaultTimeout)
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return cleaned
}

func splitCSVUpper(value string) []string {
	values := splitCSV(value)
	if len(values) == 0 {
		return nil
	}
	upper := make([]string, 0, len(values))
	for _, item := range values {
		upper = append(upper, strings.ToUpper(item))
	}
	return upper
}

func validateNextURL(next string) error {
	next = strings.TrimSpace(next)
	if next == "" {
		return nil
	}
	parsed, err := url.Parse(next)
	if err != nil {
		return fmt.Errorf("--next must be a valid URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "api.appstoreconnect.apple.com" {
		return fmt.Errorf("--next must be an App Store Connect URL")
	}
	return nil
}

func validateSort(value string, allowed ...string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, option := range allowed {
		if value == option {
			return nil
		}
	}
	return fmt.Errorf("--sort must be one of: %s", strings.Join(allowed, ", "))
}
