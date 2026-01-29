package asc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// WebhookDeleteResult represents CLI output for webhook deletions.
type WebhookDeleteResult struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

func webhookEventTypes(values []WebhookEventType) string {
	if len(values) == 0 {
		return ""
	}
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, string(value))
	}
	return strings.Join(items, ", ")
}

func printWebhooksTable(resp *WebhooksResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tEnabled\tURL\tEvents")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			strconv.FormatBool(item.Attributes.Enabled),
			compactWhitespace(item.Attributes.URL),
			compactWhitespace(webhookEventTypes(item.Attributes.EventTypes)),
		)
	}
	return w.Flush()
}

func printWebhooksMarkdown(resp *WebhooksResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Enabled | URL | Events |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Name),
			escapeMarkdown(strconv.FormatBool(item.Attributes.Enabled)),
			escapeMarkdown(item.Attributes.URL),
			escapeMarkdown(webhookEventTypes(item.Attributes.EventTypes)),
		)
	}
	return nil
}

func printWebhookDeliveriesTable(resp *WebhookDeliveriesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tState\tCreated\tSent\tError")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			compactWhitespace(item.Attributes.DeliveryState),
			compactWhitespace(item.Attributes.CreatedDate),
			compactWhitespace(item.Attributes.SentDate),
			compactWhitespace(item.Attributes.ErrorMessage),
		)
	}
	return w.Flush()
}

func printWebhookDeliveriesMarkdown(resp *WebhookDeliveriesResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | State | Created | Sent | Error |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %s | %s | %s |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.DeliveryState),
			escapeMarkdown(item.Attributes.CreatedDate),
			escapeMarkdown(item.Attributes.SentDate),
			escapeMarkdown(item.Attributes.ErrorMessage),
		)
	}
	return nil
}

func printWebhookDeleteResultTable(result *WebhookDeleteResult) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDeleted")
	fmt.Fprintf(w, "%s\t%t\n", result.ID, result.Deleted)
	return w.Flush()
}

func printWebhookDeleteResultMarkdown(result *WebhookDeleteResult) error {
	fmt.Fprintln(os.Stdout, "| ID | Deleted |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	fmt.Fprintf(os.Stdout, "| %s | %t |\n", escapeMarkdown(result.ID), result.Deleted)
	return nil
}

func printWebhookPingTable(resp *WebhookPingResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID")
	fmt.Fprintf(w, "%s\n", resp.Data.ID)
	return w.Flush()
}

func printWebhookPingMarkdown(resp *WebhookPingResponse) error {
	fmt.Fprintln(os.Stdout, "| ID |")
	fmt.Fprintln(os.Stdout, "| --- |")
	fmt.Fprintf(os.Stdout, "| %s |\n", escapeMarkdown(resp.Data.ID))
	return nil
}
