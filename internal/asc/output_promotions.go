package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// AppStoreVersionPromotionCreateResult represents CLI output for promotion creation.
type AppStoreVersionPromotionCreateResult struct {
	PromotionID string `json:"promotionId"`
	VersionID   string `json:"versionId"`
	TreatmentID string `json:"treatmentId,omitempty"`
}

func printAppStoreVersionPromotionCreateTable(result *AppStoreVersionPromotionCreateResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Promotion ID\tVersion ID\tTreatment ID")
	fmt.Fprintf(w, "%s\t%s\t%s\n", result.PromotionID, result.VersionID, result.TreatmentID)
	return w.Flush()
}

func printAppStoreVersionPromotionCreateMarkdown(result *AppStoreVersionPromotionCreateResult) error {
	fmt.Fprintln(os.Stdout, "| Promotion ID | Version ID | Treatment ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s |\n",
		escapeMarkdown(result.PromotionID),
		escapeMarkdown(result.VersionID),
		escapeMarkdown(result.TreatmentID),
	)
	return nil
}
