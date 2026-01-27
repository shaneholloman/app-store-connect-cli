package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// AppStoreVersionSubmissionResult represents CLI output for submissions.
type AppStoreVersionSubmissionResult struct {
	SubmissionID string  `json:"submissionId"`
	CreatedDate  *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionCreateResult represents CLI output for submission creation.
type AppStoreVersionSubmissionCreateResult struct {
	SubmissionID string  `json:"submissionId"`
	VersionID    string  `json:"versionId"`
	BuildID      string  `json:"buildId"`
	CreatedDate  *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionStatusResult represents CLI output for submission status.
type AppStoreVersionSubmissionStatusResult struct {
	ID            string  `json:"id"`
	VersionID     string  `json:"versionId,omitempty"`
	VersionString string  `json:"versionString,omitempty"`
	Platform      string  `json:"platform,omitempty"`
	State         string  `json:"state,omitempty"`
	CreatedDate   *string `json:"createdDate,omitempty"`
}

// AppStoreVersionSubmissionCancelResult represents CLI output for submission cancellation.
type AppStoreVersionSubmissionCancelResult struct {
	ID        string `json:"id"`
	Cancelled bool   `json:"cancelled"`
}

// AppStoreVersionDetailResult represents CLI output for version details.
type AppStoreVersionDetailResult struct {
	ID            string `json:"id"`
	VersionString string `json:"versionString,omitempty"`
	Platform      string `json:"platform,omitempty"`
	State         string `json:"state,omitempty"`
	BuildID       string `json:"buildId,omitempty"`
	BuildVersion  string `json:"buildVersion,omitempty"`
	SubmissionID  string `json:"submissionId,omitempty"`
}

// AppStoreVersionAttachBuildResult represents CLI output for build attachment.
type AppStoreVersionAttachBuildResult struct {
	VersionID string `json:"versionId"`
	BuildID   string `json:"buildId"`
	Attached  bool   `json:"attached"`
}

// AppStoreVersionReleaseRequestResult represents CLI output for release requests.
type AppStoreVersionReleaseRequestResult struct {
	ReleaseRequestID string `json:"releaseRequestId"`
	VersionID        string `json:"versionId"`
}

func printAppStoreVersionsTable(resp *AppStoreVersionsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tVersion\tPlatform\tState\tCreated")
	for _, item := range resp.Data {
		state := item.Attributes.AppVersionState
		if state == "" {
			state = item.Attributes.AppStoreState
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Attributes.VersionString,
			string(item.Attributes.Platform),
			state,
			item.Attributes.CreatedDate,
		)
	}
	return w.Flush()
}

func printPreReleaseVersionsTable(resp *PreReleaseVersionsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tVersion\tPlatform")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Version),
			string(item.Attributes.Platform),
		)
	}
	return w.Flush()
}

func printAppStoreVersionsMarkdown(resp *AppStoreVersionsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Version | Platform | State | Created |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		state := item.Attributes.AppVersionState
		if state == "" {
			state = item.Attributes.AppStoreState
		}
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.VersionString),
			escapeMarkdown(string(item.Attributes.Platform)),
			escapeMarkdown(state),
			escapeMarkdown(item.Attributes.CreatedDate),
		)
	}
	return nil
}

func printPreReleaseVersionsMarkdown(resp *PreReleaseVersionsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Version | Platform |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Version),
			escapeMarkdown(string(item.Attributes.Platform)),
		)
	}
	return nil
}

func printAppStoreVersionSubmissionTable(result *AppStoreVersionSubmissionResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tCreated Date")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(w, "%s\t%s\n", result.SubmissionID, createdDate)
	return w.Flush()
}

func printAppStoreVersionSubmissionCreateTable(result *AppStoreVersionSubmissionCreateResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tVersion ID\tBuild ID\tCreated Date")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		result.SubmissionID,
		result.VersionID,
		result.BuildID,
		createdDate,
	)
	return w.Flush()
}

func printAppStoreVersionSubmissionStatusTable(result *AppStoreVersionSubmissionStatusResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tVersion ID\tVersion\tPlatform\tState\tCreated Date")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
		result.ID,
		result.VersionID,
		result.VersionString,
		result.Platform,
		result.State,
		createdDate,
	)
	return w.Flush()
}

func printAppStoreVersionSubmissionCancelTable(result *AppStoreVersionSubmissionCancelResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Submission ID\tCancelled")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Cancelled)
	return w.Flush()
}

func printAppStoreVersionDetailTable(result *AppStoreVersionDetailResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version ID\tVersion\tPlatform\tState\tBuild ID\tBuild Version\tSubmission ID")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		result.ID,
		result.VersionString,
		result.Platform,
		result.State,
		result.BuildID,
		result.BuildVersion,
		result.SubmissionID,
	)
	return w.Flush()
}

func printAppStoreVersionPhasedReleaseTable(resp *AppStoreVersionPhasedReleaseResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Phased Release ID\tState\tStart Date\tCurrent Day\tTotal Pause Duration")
	attrs := resp.Data.Attributes
	fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
		resp.Data.ID,
		attrs.PhasedReleaseState,
		attrs.StartDate,
		attrs.CurrentDayNumber,
		attrs.TotalPauseDuration,
	)
	return w.Flush()
}

func printAppStoreVersionPhasedReleaseDeleteResultTable(result *AppStoreVersionPhasedReleaseDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Phased Release ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppStoreVersionAttachBuildTable(result *AppStoreVersionAttachBuildResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Version ID\tBuild ID\tAttached")
	fmt.Fprintf(w, "%s\t%s\t%t\n", result.VersionID, result.BuildID, result.Attached)
	return w.Flush()
}

func printAppStoreVersionReleaseRequestTable(result *AppStoreVersionReleaseRequestResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Release Request ID\tVersion ID")
	fmt.Fprintf(w, "%s\t%s\n", result.ReleaseRequestID, result.VersionID)
	return w.Flush()
}

func printAppStoreVersionSubmissionMarkdown(result *AppStoreVersionSubmissionResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(os.Stdout, "| %s | %s |\n",
		escapeMarkdown(result.SubmissionID),
		escapeMarkdown(createdDate),
	)
	return nil
}

func printAppStoreVersionSubmissionCreateMarkdown(result *AppStoreVersionSubmissionCreateResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Version ID | Build ID | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- |")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s |\n",
		escapeMarkdown(result.SubmissionID),
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.BuildID),
		escapeMarkdown(createdDate),
	)
	return nil
}

func printAppStoreVersionSubmissionStatusMarkdown(result *AppStoreVersionSubmissionStatusResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Version ID | Version | Platform | State | Created Date |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- |")
	createdDate := ""
	if result.CreatedDate != nil {
		createdDate = *result.CreatedDate
	}
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.VersionString),
		escapeMarkdown(result.Platform),
		escapeMarkdown(result.State),
		escapeMarkdown(createdDate),
	)
	return nil
}

func printAppStoreVersionSubmissionCancelMarkdown(result *AppStoreVersionSubmissionCancelResult) error {
	fmt.Fprintln(os.Stdout, "| Submission ID | Cancelled |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Cancelled,
	)
	return nil
}

func printAppStoreVersionDetailMarkdown(result *AppStoreVersionDetailResult) error {
	fmt.Fprintln(os.Stdout, "| Version ID | Version | Platform | State | Build ID | Build Version | Submission ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s | %s | %s |\n",
		escapeMarkdown(result.ID),
		escapeMarkdown(result.VersionString),
		escapeMarkdown(result.Platform),
		escapeMarkdown(result.State),
		escapeMarkdown(result.BuildID),
		escapeMarkdown(result.BuildVersion),
		escapeMarkdown(result.SubmissionID),
	)
	return nil
}

func printAppStoreVersionPhasedReleaseMarkdown(resp *AppStoreVersionPhasedReleaseResponse) error {
	fmt.Fprintln(os.Stdout, "| Phased Release ID | State | Start Date | Current Day | Total Pause Duration |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	attrs := resp.Data.Attributes
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %d | %d |\n",
		escapeMarkdown(resp.Data.ID),
		escapeMarkdown(string(attrs.PhasedReleaseState)),
		escapeMarkdown(attrs.StartDate),
		attrs.CurrentDayNumber,
		attrs.TotalPauseDuration,
	)
	return nil
}

func printAppStoreVersionPhasedReleaseDeleteResultMarkdown(result *AppStoreVersionPhasedReleaseDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| Phased Release ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n",
		escapeMarkdown(result.ID),
		result.Deleted,
	)
	return nil
}

func printAppStoreVersionAttachBuildMarkdown(result *AppStoreVersionAttachBuildResult) error {
	fmt.Fprintln(os.Stdout, "| Version ID | Build ID | Attached |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %t |\n",
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.BuildID),
		result.Attached,
	)
	return nil
}

func printAppStoreVersionReleaseRequestMarkdown(result *AppStoreVersionReleaseRequestResult) error {
	fmt.Fprintln(os.Stdout, "| Release Request ID | Version ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s |\n",
		escapeMarkdown(result.ReleaseRequestID),
		escapeMarkdown(result.VersionID),
	)
	return nil
}
