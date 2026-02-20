package insights

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
	"time"
)

func TestNormalizeWeekStart(t *testing.T) {
	parsed, err := normalizeWeekStart("2026-02-16")
	if err != nil {
		t.Fatalf("normalizeWeekStart error: %v", err)
	}
	if parsed.Format("2006-01-02") != "2026-02-16" {
		t.Fatalf("unexpected week start %q", parsed.Format("2006-01-02"))
	}

	if _, err := normalizeWeekStart("2026-2-16"); err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestParseSalesReportMetrics(t *testing.T) {
	report := strings.Join([]string{
		"Provider\tUnits\tDeveloper Proceeds\tCustomer Price",
		"foo\t10\t2.50\t4.99",
		"foo\t3\t0.75\t1.99",
		"",
	}, "\n")

	compressed := gzipText(t, report)
	metrics, err := parseSalesReportMetrics(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("parseSalesReportMetrics error: %v", err)
	}

	if metrics.rowCount != 2 {
		t.Fatalf("expected rowCount=2, got %d", metrics.rowCount)
	}
	if !metrics.hasUnits || metrics.unitsTotal != 13 {
		t.Fatalf("unexpected units totals: %+v", metrics)
	}
	if !metrics.hasDeveloperProceeds || metrics.developerProceedsTotal != 3.25 {
		t.Fatalf("unexpected developer proceeds totals: %+v", metrics)
	}
	if !metrics.hasCustomerPrice || metrics.customerPriceTotal != 6.98 {
		t.Fatalf("unexpected customer price totals: %+v", metrics)
	}
}

func TestContainsDate(t *testing.T) {
	window := weekWindowFromStart(time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC))

	if !containsDate(window, time.Date(2026, 2, 16, 15, 0, 0, 0, time.UTC)) {
		t.Fatal("expected first day to be in range")
	}
	if !containsDate(window, time.Date(2026, 2, 22, 23, 59, 0, 0, time.UTC)) {
		t.Fatal("expected last day to be in range")
	}
	if containsDate(window, time.Date(2026, 2, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("expected next week date to be out of range")
	}
}

func gzipText(t *testing.T, value string) []byte {
	t.Helper()

	var out bytes.Buffer
	zw := gzip.NewWriter(&out)
	if _, err := zw.Write([]byte(value)); err != nil {
		t.Fatalf("gzip write error: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close error: %v", err)
	}
	return out.Bytes()
}
