package web

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/cli/shared"
	webcore "github.com/rudrankriyam/App-Store-Connect-CLI/internal/web"
)

type reviewAttachmentDownloadResult struct {
	AttachmentID      string `json:"attachmentId"`
	SourceType        string `json:"sourceType"`
	FileName          string `json:"fileName"`
	Path              string `json:"path"`
	ThreadID          string `json:"threadId,omitempty"`
	MessageID         string `json:"messageId,omitempty"`
	ReviewRejectionID string `json:"reviewRejectionId,omitempty"`
	RefreshedURL      bool   `json:"refreshedUrl,omitempty"`
}

func validateExactlyOneSelector(valueA, flagA, valueB, flagB string) error {
	hasA := strings.TrimSpace(valueA) != ""
	hasB := strings.TrimSpace(valueB) != ""
	if hasA == hasB {
		return shared.UsageErrorf("exactly one of --%s or --%s is required", flagA, flagB)
	}
	return nil
}

func listReviewAttachmentsForSelector(
	ctx context.Context,
	client *webcore.Client,
	threadID, submissionID string,
	includeURL bool,
) ([]webcore.ReviewAttachment, error) {
	if strings.TrimSpace(threadID) != "" {
		return client.ListReviewAttachmentsByThread(ctx, threadID, includeURL)
	}
	return client.ListReviewAttachmentsBySubmission(ctx, submissionID, includeURL)
}

func normalizeAttachmentFilename(attachment webcore.ReviewAttachment) string {
	name := strings.TrimSpace(attachment.FileName)
	if name != "" {
		base := filepath.Base(name)
		if base != "" && base != "." && base != string(filepath.Separator) && base != ".." {
			return base
		}
	}
	id := strings.TrimSpace(attachment.AttachmentID)
	if id == "" {
		id = "attachment"
	}
	return id + ".bin"
}

func resolveDownloadPath(outDir, fileName string, overwrite bool) (string, error) {
	base := filepath.Join(outDir, fileName)
	if overwrite {
		return base, nil
	}
	if _, err := os.Stat(base); err == nil {
		ext := filepath.Ext(fileName)
		stem := strings.TrimSuffix(fileName, ext)
		if stem == "" {
			stem = "attachment"
		}
		for i := 1; i <= 10_000; i++ {
			candidate := filepath.Join(outDir, fmt.Sprintf("%s-%d%s", stem, i, ext))
			if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
				return candidate, nil
			}
		}
		return "", fmt.Errorf("failed to generate unique filename for %q", fileName)
	} else if errors.Is(err, os.ErrNotExist) {
		return base, nil
	} else {
		return "", fmt.Errorf("failed to check destination path %q: %w", base, err)
	}
}

func attachmentRefreshKey(attachment webcore.ReviewAttachment) string {
	return strings.Join([]string{
		strings.TrimSpace(attachment.SourceType),
		strings.TrimSpace(attachment.AttachmentID),
		strings.TrimSpace(attachment.ThreadID),
		strings.TrimSpace(attachment.MessageID),
		strings.TrimSpace(attachment.ReviewRejectionID),
	}, "|")
}

func indexAttachmentsByRefreshKey(attachments []webcore.ReviewAttachment) map[string]webcore.ReviewAttachment {
	result := make(map[string]webcore.ReviewAttachment, len(attachments))
	for _, attachment := range attachments {
		result[attachmentRefreshKey(attachment)] = attachment
	}
	return result
}

func attachmentDownloadResult(attachment webcore.ReviewAttachment, path string, refreshed bool) reviewAttachmentDownloadResult {
	return reviewAttachmentDownloadResult{
		AttachmentID:      attachment.AttachmentID,
		SourceType:        attachment.SourceType,
		FileName:          normalizeAttachmentFilename(attachment),
		Path:              path,
		ThreadID:          attachment.ThreadID,
		MessageID:         attachment.MessageID,
		ReviewRejectionID: attachment.ReviewRejectionID,
		RefreshedURL:      refreshed,
	}
}

// WebReviewCommand returns the detached web review traversal command group.
func WebReviewCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review", flag.ExitOnError)

	return &ffcli.Command{
		Name:       "review",
		ShortUsage: "asc web review <subcommand> [flags]",
		ShortHelp:  "EXPERIMENTAL: Resolution Center data via unofficial web APIs.",
		LongHelp: `EXPERIMENTAL / UNOFFICIAL / DISCOURAGED

Traverse App Review issues, messages, rejections, and attachments through web-session endpoints.

` + webWarningText,
		FlagSet:   fs,
		UsageFunc: shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			WebReviewThreadsCommand(),
			WebReviewMessagesCommand(),
			WebReviewRejectionsCommand(),
			WebReviewDraftCommand(),
			WebReviewAttachmentsCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// WebReviewThreadsCommand contains thread traversal subcommands.
func WebReviewThreadsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review threads", flag.ExitOnError)
	return &ffcli.Command{
		Name:        "threads",
		ShortUsage:  "asc web review threads <subcommand> [flags]",
		ShortHelp:   "EXPERIMENTAL: Resolution Center thread access.",
		FlagSet:     fs,
		UsageFunc:   shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{WebReviewThreadsListCommand()},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// WebReviewThreadsListCommand lists threads by app or submission.
func WebReviewThreadsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review threads list", flag.ExitOnError)
	appID := fs.String("app", "", "App ID")
	submissionID := fs.String("submission", "", "Review submission ID")
	authFlags := bindWebSessionFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc web review threads list (--app APP_ID | --submission REVIEW_SUBMISSION_ID) [flags]",
		ShortHelp:  "EXPERIMENTAL: List Resolution Center threads.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedAppID := strings.TrimSpace(*appID)
			trimmedSubmissionID := strings.TrimSpace(*submissionID)
			if err := validateExactlyOneSelector(trimmedAppID, "app", trimmedSubmissionID, "submission"); err != nil {
				return err
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			session, err := resolveWebSessionForCommand(requestCtx, authFlags)
			if err != nil {
				return err
			}
			client := webcore.NewClient(session)

			var threads []webcore.ResolutionCenterThread
			if trimmedAppID != "" {
				threads, err = client.ListResolutionCenterThreadsByApp(requestCtx, trimmedAppID)
			} else {
				threads, err = client.ListResolutionCenterThreadsBySubmission(requestCtx, trimmedSubmissionID)
			}
			if err != nil {
				return withWebAuthHint(err, "web review threads list")
			}
			return shared.PrintOutput(threads, *output.Output, *output.Pretty)
		},
	}
}

// WebReviewMessagesCommand contains message traversal subcommands.
func WebReviewMessagesCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review messages", flag.ExitOnError)
	return &ffcli.Command{
		Name:        "messages",
		ShortUsage:  "asc web review messages <subcommand> [flags]",
		ShortHelp:   "EXPERIMENTAL: Resolution Center message access.",
		FlagSet:     fs,
		UsageFunc:   shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{WebReviewMessagesListCommand()},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// WebReviewMessagesListCommand lists messages for a thread.
func WebReviewMessagesListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review messages list", flag.ExitOnError)
	threadID := fs.String("thread", "", "Resolution Center thread ID")
	plainText := fs.Bool("plain-text", false, "Project messageBody HTML into plain text")
	authFlags := bindWebSessionFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc web review messages list --thread THREAD_ID [--plain-text] [flags]",
		ShortHelp:  "EXPERIMENTAL: List Resolution Center messages.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedThreadID := strings.TrimSpace(*threadID)
			if trimmedThreadID == "" {
				return shared.UsageError("--thread is required")
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			session, err := resolveWebSessionForCommand(requestCtx, authFlags)
			if err != nil {
				return err
			}
			client := webcore.NewClient(session)

			messages, err := client.ListResolutionCenterMessages(requestCtx, trimmedThreadID, *plainText)
			if err != nil {
				return withWebAuthHint(err, "web review messages list")
			}
			return shared.PrintOutput(messages, *output.Output, *output.Pretty)
		},
	}
}

// WebReviewRejectionsCommand contains rejection traversal subcommands.
func WebReviewRejectionsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review rejections", flag.ExitOnError)
	return &ffcli.Command{
		Name:        "rejections",
		ShortUsage:  "asc web review rejections <subcommand> [flags]",
		ShortHelp:   "EXPERIMENTAL: Review rejection access.",
		FlagSet:     fs,
		UsageFunc:   shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{WebReviewRejectionsListCommand()},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// WebReviewRejectionsListCommand lists rejections for a thread.
func WebReviewRejectionsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review rejections list", flag.ExitOnError)
	threadID := fs.String("thread", "", "Resolution Center thread ID")
	authFlags := bindWebSessionFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc web review rejections list --thread THREAD_ID [flags]",
		ShortHelp:  "EXPERIMENTAL: List review rejection reasons.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedThreadID := strings.TrimSpace(*threadID)
			if trimmedThreadID == "" {
				return shared.UsageError("--thread is required")
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			session, err := resolveWebSessionForCommand(requestCtx, authFlags)
			if err != nil {
				return err
			}
			client := webcore.NewClient(session)

			rejections, err := client.ListReviewRejections(requestCtx, trimmedThreadID)
			if err != nil {
				return withWebAuthHint(err, "web review rejections list")
			}
			return shared.PrintOutput(rejections, *output.Output, *output.Pretty)
		},
	}
}

// WebReviewDraftCommand contains draft traversal subcommands.
func WebReviewDraftCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review draft", flag.ExitOnError)
	return &ffcli.Command{
		Name:        "draft",
		ShortUsage:  "asc web review draft <subcommand> [flags]",
		ShortHelp:   "EXPERIMENTAL: Draft message access.",
		FlagSet:     fs,
		UsageFunc:   shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{WebReviewDraftShowCommand()},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// WebReviewDraftShowCommand shows draft payload for a thread.
func WebReviewDraftShowCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review draft show", flag.ExitOnError)
	threadID := fs.String("thread", "", "Resolution Center thread ID")
	plainText := fs.Bool("plain-text", false, "Project messageBody HTML into plain text")
	authFlags := bindWebSessionFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "show",
		ShortUsage: "asc web review draft show --thread THREAD_ID [--plain-text] [flags]",
		ShortHelp:  "EXPERIMENTAL: Show draft message for a thread.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedThreadID := strings.TrimSpace(*threadID)
			if trimmedThreadID == "" {
				return shared.UsageError("--thread is required")
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			session, err := resolveWebSessionForCommand(requestCtx, authFlags)
			if err != nil {
				return err
			}
			client := webcore.NewClient(session)

			draft, err := client.GetResolutionCenterDraftMessage(requestCtx, trimmedThreadID, *plainText)
			if err != nil {
				return withWebAuthHint(err, "web review draft show")
			}
			if draft == nil {
				return shared.PrintOutput(map[string]any{}, *output.Output, *output.Pretty)
			}
			return shared.PrintOutput(draft, *output.Output, *output.Pretty)
		},
	}
}

// WebReviewAttachmentsCommand contains attachment list/download subcommands.
func WebReviewAttachmentsCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review attachments", flag.ExitOnError)
	return &ffcli.Command{
		Name:       "attachments",
		ShortUsage: "asc web review attachments <subcommand> [flags]",
		ShortHelp:  "EXPERIMENTAL: Review attachment listing and download.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Subcommands: []*ffcli.Command{
			WebReviewAttachmentsListCommand(),
			WebReviewAttachmentsDownloadCommand(),
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}
}

// WebReviewAttachmentsListCommand lists attachments by thread or submission.
func WebReviewAttachmentsListCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review attachments list", flag.ExitOnError)
	threadID := fs.String("thread", "", "Resolution Center thread ID")
	submissionID := fs.String("submission", "", "Review submission ID")
	includeURL := fs.Bool("include-url", false, "Include signed downloadUrl in output (sensitive)")
	authFlags := bindWebSessionFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "list",
		ShortUsage: "asc web review attachments list (--thread THREAD_ID | --submission REVIEW_SUBMISSION_ID) [--include-url] [flags]",
		ShortHelp:  "EXPERIMENTAL: List review attachments.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedThreadID := strings.TrimSpace(*threadID)
			trimmedSubmissionID := strings.TrimSpace(*submissionID)
			if err := validateExactlyOneSelector(trimmedThreadID, "thread", trimmedSubmissionID, "submission"); err != nil {
				return err
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			session, err := resolveWebSessionForCommand(requestCtx, authFlags)
			if err != nil {
				return err
			}
			client := webcore.NewClient(session)

			attachments, err := listReviewAttachmentsForSelector(
				requestCtx,
				client,
				trimmedThreadID,
				trimmedSubmissionID,
				*includeURL,
			)
			if err != nil {
				return withWebAuthHint(err, "web review attachments list")
			}
			return shared.PrintOutput(attachments, *output.Output, *output.Pretty)
		},
	}
}

// WebReviewAttachmentsDownloadCommand downloads attachments by thread or submission.
func WebReviewAttachmentsDownloadCommand() *ffcli.Command {
	fs := flag.NewFlagSet("web review attachments download", flag.ExitOnError)
	threadID := fs.String("thread", "", "Resolution Center thread ID")
	submissionID := fs.String("submission", "", "Review submission ID")
	outDir := fs.String("out", "", "Output directory for downloaded files")
	pattern := fs.String("pattern", "", "Optional filename glob filter (for example: *.png)")
	overwrite := fs.Bool("overwrite", false, "Overwrite existing files instead of suffixing")
	authFlags := bindWebSessionFlags(fs)
	output := shared.BindOutputFlags(fs)

	return &ffcli.Command{
		Name:       "download",
		ShortUsage: "asc web review attachments download (--thread THREAD_ID | --submission REVIEW_SUBMISSION_ID) --out DIR [--pattern GLOB] [--overwrite] [flags]",
		ShortHelp:  "EXPERIMENTAL: Download review attachments.",
		FlagSet:    fs,
		UsageFunc:  shared.DefaultUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			trimmedThreadID := strings.TrimSpace(*threadID)
			trimmedSubmissionID := strings.TrimSpace(*submissionID)
			if err := validateExactlyOneSelector(trimmedThreadID, "thread", trimmedSubmissionID, "submission"); err != nil {
				return err
			}
			trimmedOutDir := strings.TrimSpace(*outDir)
			if trimmedOutDir == "" {
				return shared.UsageError("--out is required")
			}
			if strings.TrimSpace(*pattern) != "" {
				if _, err := filepath.Match(strings.TrimSpace(*pattern), "sample.png"); err != nil {
					return shared.UsageErrorf("--pattern is invalid: %v", err)
				}
			}

			requestCtx, cancel := shared.ContextWithTimeout(ctx)
			defer cancel()

			session, err := resolveWebSessionForCommand(requestCtx, authFlags)
			if err != nil {
				return err
			}
			client := webcore.NewClient(session)

			attachments, err := listReviewAttachmentsForSelector(
				requestCtx,
				client,
				trimmedThreadID,
				trimmedSubmissionID,
				true,
			)
			if err != nil {
				return withWebAuthHint(err, "web review attachments download")
			}

			if err := os.MkdirAll(trimmedOutDir, 0o755); err != nil {
				return fmt.Errorf("failed to create output directory %q: %w", trimmedOutDir, err)
			}

			selected := make([]webcore.ReviewAttachment, 0, len(attachments))
			for _, attachment := range attachments {
				attachment.FileName = normalizeAttachmentFilename(attachment)
				if !attachment.Downloadable || strings.TrimSpace(attachment.DownloadURL) == "" {
					continue
				}
				if strings.TrimSpace(*pattern) != "" {
					matched, err := filepath.Match(strings.TrimSpace(*pattern), attachment.FileName)
					if err != nil {
						return shared.UsageErrorf("--pattern is invalid: %v", err)
					}
					if !matched {
						continue
					}
				}
				selected = append(selected, attachment)
			}

			results := make([]reviewAttachmentDownloadResult, 0, len(selected))
			failures := make([]string, 0)
			var refreshedIndex map[string]webcore.ReviewAttachment

			for _, attachment := range selected {
				body, statusCode, downloadErr := client.DownloadAttachment(requestCtx, attachment.DownloadURL)
				refreshed := false
				if downloadErr != nil && (statusCode == http.StatusForbidden || statusCode == http.StatusGone) {
					if refreshedIndex == nil {
						refreshedAttachments, refreshErr := listReviewAttachmentsForSelector(
							requestCtx,
							client,
							trimmedThreadID,
							trimmedSubmissionID,
							true,
						)
						if refreshErr != nil {
							failures = append(
								failures,
								fmt.Sprintf("%s: refresh failed (%v)", attachment.FileName, refreshErr),
							)
							continue
						}
						refreshedIndex = indexAttachmentsByRefreshKey(refreshedAttachments)
					}
					if refreshedAttachment, ok := refreshedIndex[attachmentRefreshKey(attachment)]; ok && strings.TrimSpace(refreshedAttachment.DownloadURL) != "" {
						body, _, downloadErr = client.DownloadAttachment(requestCtx, refreshedAttachment.DownloadURL)
						if downloadErr == nil {
							attachment = refreshedAttachment
							attachment.FileName = normalizeAttachmentFilename(attachment)
							refreshed = true
						}
					}
				}
				if downloadErr != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", attachment.FileName, downloadErr))
					continue
				}

				outputPath, err := resolveDownloadPath(trimmedOutDir, attachment.FileName, *overwrite)
				if err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", attachment.FileName, err))
					continue
				}
				if err := os.WriteFile(outputPath, body, 0o600); err != nil {
					failures = append(failures, fmt.Sprintf("%s: %v", attachment.FileName, err))
					continue
				}
				results = append(results, attachmentDownloadResult(attachment, outputPath, refreshed))
			}

			if err := shared.PrintOutput(results, *output.Output, *output.Pretty); err != nil {
				return err
			}
			if len(failures) > 0 {
				return fmt.Errorf("download failed for %d attachment(s): %s", len(failures), strings.Join(failures, "; "))
			}
			return nil
		},
	}
}
