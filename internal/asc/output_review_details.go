package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func formatReviewDetailContactName(attr AppStoreReviewDetailAttributes) string {
	first := strings.TrimSpace(attr.ContactFirstName)
	last := strings.TrimSpace(attr.ContactLastName)
	switch {
	case first == "" && last == "":
		return ""
	case first == "":
		return last
	case last == "":
		return first
	default:
		return first + " " + last
	}
}

func printAppStoreReviewDetailTable(resp *AppStoreReviewDetailResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tContact\tEmail\tPhone\tDemo Required\tDemo Account\tNotes")
	attr := resp.Data.Attributes
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%s\t%s\n",
		resp.Data.ID,
		compactWhitespace(formatReviewDetailContactName(attr)),
		compactWhitespace(attr.ContactEmail),
		compactWhitespace(attr.ContactPhone),
		attr.DemoAccountRequired,
		compactWhitespace(attr.DemoAccountName),
		compactWhitespace(attr.Notes),
	)
	return w.Flush()
}

func printAppStoreReviewDetailMarkdown(resp *AppStoreReviewDetailResponse) error {
	attr := resp.Data.Attributes
	fmt.Fprintln(os.Stdout, "| ID | Contact | Email | Phone | Demo Required | Demo Account | Notes |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- | --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %t | %s | %s |\n",
		escapeMarkdown(resp.Data.ID),
		escapeMarkdown(formatReviewDetailContactName(attr)),
		escapeMarkdown(attr.ContactEmail),
		escapeMarkdown(attr.ContactPhone),
		attr.DemoAccountRequired,
		escapeMarkdown(attr.DemoAccountName),
		escapeMarkdown(attr.Notes),
	)
	return nil
}
