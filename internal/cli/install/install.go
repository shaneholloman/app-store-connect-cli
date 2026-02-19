package install

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
)

const defaultSkillsPackage = "rudrankriyam/asc-skills"

var (
	lookupNpx      = exec.LookPath
	runCommand     = defaultRunCommand
	errNpxNotFound = errors.New("npx not found")
	validPackage   = regexp.MustCompile(`^[A-Za-z0-9@._/-]+$`)
)

// InstallCommand returns the install command factory.
func InstallCommand() *ffcli.Command {
	fs := flag.NewFlagSet("install", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "install",
		ShortUsage: "asc install <subcommand> [flags]",
		ShortHelp:  "Install optional asc components.",
		LongHelp: `Install optional asc components.

Subcommands:
  skills    Install the asc skill pack for App Store Connect workflows.`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			InstallSkillsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) == 0 {
				return flag.ErrHelp
			}
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", args[0])
			return flag.ErrHelp
		},
	}
}

// InstallSkillsCommand returns the install skills subcommand.
func InstallSkillsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("install skills", flag.ExitOnError)
	packageName := fs.String("package", defaultSkillsPackage, "NPM package name or repo for the skill pack")

	return &ffcli.Command{
		Name:       "skills",
		ShortUsage: "asc install skills [flags]",
		ShortHelp:  "Install the asc skill pack for App Store Connect workflows.",
		LongHelp: `Install the asc skill pack for App Store Connect workflows.

Examples:
  asc install skills
  asc install skills --package "rudrankriyam/asc-skills"`,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if err := installSkills(ctx, *packageName); err != nil {
				return fmt.Errorf("install skills: %w", err)
			}
			return nil
		},
	}
}

func installSkills(ctx context.Context, pkg string) error {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return fmt.Errorf("--package is required")
	}
	if err := validatePackageName(pkg); err != nil {
		return err
	}

	path, err := lookupNpx("npx")
	if err != nil {
		return fmt.Errorf("%w; install Node.js to continue", errNpxNotFound)
	}

	return runCommand(ctx, path, "--yes", "add-skill", pkg)
}

func validatePackageName(pkg string) error {
	if strings.HasPrefix(pkg, "-") {
		return fmt.Errorf("--package must not start with '-'")
	}
	if !validPackage.MatchString(pkg) {
		return fmt.Errorf("--package must be a valid npm package or repo (letters, numbers, @, ., _, -, /)")
	}
	return nil
}

func defaultRunCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
