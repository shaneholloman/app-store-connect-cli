package webhooks

import (
	"fmt"
	"strings"

	"github.com/rudrankriyam/App-Store-Connect-CLI/internal/asc"
)

func normalizeWebhookEvents(value string) ([]asc.WebhookEventType, error) {
	values := splitCSV(value)
	if len(values) == 0 {
		return nil, fmt.Errorf("--events must include at least one value")
	}

	normalized := make([]asc.WebhookEventType, 0, len(values))
	for _, item := range values {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, asc.WebhookEventType(strings.ToUpper(trimmed)))
	}

	if len(normalized) == 0 {
		return nil, fmt.Errorf("--events must include at least one value")
	}

	return normalized, nil
}
