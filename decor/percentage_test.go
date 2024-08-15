package decor

import (
	"fmt"
	"testing"
)

func TestPercentageType(t *testing.T) {
	cases := map[string]struct {
		value    float64
		verb     string
		expected string
	}{
		"10 %d":   {10, "%d", "10%"},
		"10 %s":   {10, "%s", "10%"},
		"10 %f":   {10, "%f", "10.000000%"},
		"10 %.6f": {10, "%.6f", "10.000000%"},
		"10 %.0f": {10, "%.0f", "10%"},
		"10 %.1f": {10, "%.1f", "10.0%"},
		"10 %.2f": {10, "%.2f", "10.00%"},
		"10 %.3f": {10, "%.3f", "10.000%"},

		"10 % d":   {10, "% d", "10 %"},
		"10 % s":   {10, "% s", "10 %"},
		"10 % f":   {10, "% f", "10.000000 %"},
		"10 % .6f": {10, "% .6f", "10.000000 %"},
		"10 % .0f": {10, "% .0f", "10 %"},
		"10 % .1f": {10, "% .1f", "10.0 %"},
		"10 % .2f": {10, "% .2f", "10.00 %"},
		"10 % .3f": {10, "% .3f", "10.000 %"},

		"10.5 %d":   {10.5, "%d", "10%"},
		"10.5 %s":   {10.5, "%s", "10%"},
		"10.5 %f":   {10.5, "%f", "10.500000%"},
		"10.5 %.6f": {10.5, "%.6f", "10.500000%"},
		"10.5 %.0f": {10.5, "%.0f", "10%"},
		"10.5 %.1f": {10.5, "%.1f", "10.5%"},
		"10.5 %.2f": {10.5, "%.2f", "10.50%"},
		"10.5 %.3f": {10.5, "%.3f", "10.500%"},

		"10.5 % d":   {10.5, "% d", "10 %"},
		"10.5 % s":   {10.5, "% s", "10 %"},
		"10.5 % f":   {10.5, "% f", "10.500000 %"},
		"10.5 % .6f": {10.5, "% .6f", "10.500000 %"},
		"10.5 % .0f": {10.5, "% .0f", "10 %"},
		"10.5 % .1f": {10.5, "% .1f", "10.5 %"},
		"10.5 % .2f": {10.5, "% .2f", "10.50 %"},
		"10.5 % .3f": {10.5, "% .3f", "10.500 %"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fmt.Sprintf(tc.verb, percentageType(tc.value))
			if got != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, got)
			}
		})
	}
}

func TestPercentageDecor(t *testing.T) {
	cases := []struct {
		name     string
		fmt      string
		current  int64
		total    int64
		expected string
	}{
		{
			name:     "tot:100 cur:0 fmt:none",
			fmt:      "",
			current:  0,
			total:    100,
			expected: "0 %",
		},
		{
			name:     "tot:100 cur:10 fmt:none",
			fmt:      "",
			current:  10,
			total:    100,
			expected: "10 %",
		},
		{
			name:     "tot:100 cur:10 fmt:%.2f",
			fmt:      "%.2f",
			current:  10,
			total:    100,
			expected: "10.00%",
		},
		{
			name:     "tot:99 cur:10 fmt:%.2f",
			fmt:      "%.2f",
			current:  11,
			total:    99,
			expected: "11.11%",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decor := NewPercentage(tc.fmt)
			stat := Statistics{
				Total:   tc.total,
				Current: tc.current,
			}
			res, _ := decor.Decor(stat)
			if res != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, res)
			}
		})
	}
}
