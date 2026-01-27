package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type appStoreReviewAttachmentField struct {
	Name  string
	Value string
}

func printAppStoreReviewAttachmentsTable(resp *AppStoreReviewAttachmentsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFile Name\tFile Size\tChecksum\tDelivery State")
	for _, item := range resp.Data {
		attrs := item.Attributes
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			sanitizeTerminal(item.ID),
			sanitizeTerminal(fallbackValue(attrs.FileName)),
			formatAttachmentFileSize(attrs.FileSize),
			sanitizeTerminal(fallbackValue(attrs.SourceFileChecksum)),
			sanitizeTerminal(formatAssetDeliveryState(attrs.AssetDeliveryState)),
		)
	}
	return w.Flush()
}

func printAppStoreReviewAttachmentsMarkdown(resp *AppStoreReviewAttachmentsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | File Name | File Size | Checksum | Delivery State |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		attrs := item.Attributes
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(fallbackValue(attrs.FileName)),
			escapeMarkdown(formatAttachmentFileSize(attrs.FileSize)),
			escapeMarkdown(fallbackValue(attrs.SourceFileChecksum)),
			escapeMarkdown(formatAssetDeliveryState(attrs.AssetDeliveryState)),
		)
	}
	return nil
}

func printAppStoreReviewAttachmentTable(resp *AppStoreReviewAttachmentResponse) error {
	fields := appStoreReviewAttachmentFields(resp)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Field\tValue")
	for _, field := range fields {
		fmt.Fprintf(w, "%s\t%s\n", field.Name, field.Value)
	}
	return w.Flush()
}

func printAppStoreReviewAttachmentMarkdown(resp *AppStoreReviewAttachmentResponse) error {
	fields := appStoreReviewAttachmentFields(resp)
	fmt.Fprintln(os.Stdout, "| Field | Value |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, field := range fields {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n", escapeMarkdown(field.Name), escapeMarkdown(field.Value))
	}
	return nil
}

func appStoreReviewAttachmentFields(resp *AppStoreReviewAttachmentResponse) []appStoreReviewAttachmentField {
	if resp == nil {
		return nil
	}
	attrs := resp.Data.Attributes
	return []appStoreReviewAttachmentField{
		{Name: "ID", Value: fallbackValue(resp.Data.ID)},
		{Name: "Type", Value: fallbackValue(string(resp.Data.Type))},
		{Name: "File Name", Value: fallbackValue(attrs.FileName)},
		{Name: "File Size", Value: formatAttachmentFileSize(attrs.FileSize)},
		{Name: "Source File Checksum", Value: fallbackValue(attrs.SourceFileChecksum)},
		{Name: "Delivery State", Value: formatAssetDeliveryState(attrs.AssetDeliveryState)},
	}
}

func printAppStoreReviewAttachmentDeleteResultTable(result *AppStoreReviewAttachmentDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printAppStoreReviewAttachmentDeleteResultMarkdown(result *AppStoreReviewAttachmentDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n", escapeMarkdown(result.ID), result.Deleted)
	return nil
}

func formatAssetDeliveryState(state *AppMediaAssetState) string {
	if state == nil || state.State == nil {
		return ""
	}
	return *state.State
}

func formatAttachmentFileSize(size int64) string {
	if size <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", size)
}
