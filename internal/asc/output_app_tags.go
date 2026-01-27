package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func printAppTagsTable(resp *AppTagsResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tVisible In App Store")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\t%t\n",
			item.ID,
			compactWhitespace(item.Attributes.Name),
			item.Attributes.VisibleInAppStore,
		)
	}
	return w.Flush()
}

func printAppTagsMarkdown(resp *AppTagsResponse) error {
	fmt.Fprintln(os.Stdout, "| ID | Name | Visible In App Store |")
	fmt.Fprintln(os.Stdout, "| --- | --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s | %t |\n",
			escapeMarkdown(item.ID),
			escapeMarkdown(item.Attributes.Name),
			item.Attributes.VisibleInAppStore,
		)
	}
	return nil
}
