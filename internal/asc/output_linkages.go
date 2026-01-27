package asc

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func printLinkagesTable(resp *LinkagesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Type\tID")
	for _, item := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", item.Type, item.ID)
	}
	return w.Flush()
}

func printLinkagesMarkdown(resp *LinkagesResponse) error {
	fmt.Fprintln(os.Stdout, "| Type | ID |")
	fmt.Fprintln(os.Stdout, "| --- | --- |")
	for _, item := range resp.Data {
		fmt.Fprintf(os.Stdout, "| %s | %s |\n",
			escapeMarkdown(string(item.Type)),
			escapeMarkdown(item.ID),
		)
	}
	return nil
}
