package decor

import (
	"fmt"
	"testing"
)

func TestB1024(t *testing.T) {
	cases := map[string]struct {
		value    int64
		verb     string
		expected string
	}{
		"verb %f":   {12345678, "%f", "11.773756MiB"},
		"verb %.0f": {12345678, "%.0f", "12MiB"},
		"verb %.1f": {12345678, "%.1f", "11.8MiB"},
		"verb %.2f": {12345678, "%.2f", "11.77MiB"},
		"verb %.3f": {12345678, "%.3f", "11.774MiB"},

		"verb % f":   {12345678, "% f", "11.773756 MiB"},
		"verb % .0f": {12345678, "% .0f", "12 MiB"},
		"verb % .1f": {12345678, "% .1f", "11.8 MiB"},
		"verb % .2f": {12345678, "% .2f", "11.77 MiB"},
		"verb % .3f": {12345678, "% .3f", "11.774 MiB"},

		"1000 %f":           {1000, "%f", "1000.000000b"},
		"1000 %d":           {1000, "%d", "1000b"},
		"1000 %s":           {1000, "%s", "1000b"},
		"1024 %f":           {1024, "%f", "1.000000KiB"},
		"1024 %d":           {1024, "%d", "1KiB"},
		"1024 %.1f":         {1024, "%.1f", "1.0KiB"},
		"1024 %s":           {1024, "%s", "1KiB"},
		"3*MiB+140KiB %f":   {3*int64(_iMiB) + 140*int64(_iKiB), "%f", "3.136719MiB"},
		"3*MiB+140KiB %d":   {3*int64(_iMiB) + 140*int64(_iKiB), "%d", "3MiB"},
		"3*MiB+140KiB %.1f": {3*int64(_iMiB) + 140*int64(_iKiB), "%.1f", "3.1MiB"},
		"3*MiB+140KiB %s":   {3*int64(_iMiB) + 140*int64(_iKiB), "%s", "3.13671875MiB"},
		"2*GiB %f":          {2 * int64(_iGiB), "%f", "2.000000GiB"},
		"2*GiB %d":          {2 * int64(_iGiB), "%d", "2GiB"},
		"2*GiB %.1f":        {2 * int64(_iGiB), "%.1f", "2.0GiB"},
		"2*GiB %s":          {2 * int64(_iGiB), "%s", "2GiB"},
		"4*TiB %f":          {4 * int64(_iTiB), "%f", "4.000000TiB"},
		"4*TiB %d":          {4 * int64(_iTiB), "%d", "4TiB"},
		"4*TiB %.1f":        {4 * int64(_iTiB), "%.1f", "4.0TiB"},
		"4*TiB %s":          {4 * int64(_iTiB), "%s", "4TiB"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fmt.Sprintf(tc.verb, SizeB1024(tc.value))
			if got != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, got)
			}
		})
	}
}

func TestB1000(t *testing.T) {
	cases := map[string]struct {
		value    int64
		verb     string
		expected string
	}{
		"verb %f":   {12345678, "%f", "12.345678MB"},
		"verb %.0f": {12345678, "%.0f", "12MB"},
		"verb %.1f": {12345678, "%.1f", "12.3MB"},
		"verb %.2f": {12345678, "%.2f", "12.35MB"},
		"verb %.3f": {12345678, "%.3f", "12.346MB"},

		"verb % f":   {12345678, "% f", "12.345678 MB"},
		"verb % .0f": {12345678, "% .0f", "12 MB"},
		"verb % .1f": {12345678, "% .1f", "12.3 MB"},
		"verb % .2f": {12345678, "% .2f", "12.35 MB"},
		"verb % .3f": {12345678, "% .3f", "12.346 MB"},

		"1000 %f":          {1000, "%f", "1.000000KB"},
		"1000 %d":          {1000, "%d", "1KB"},
		"1000 %s":          {1000, "%s", "1KB"},
		"1024 %f":          {1024, "%f", "1.024000KB"},
		"1024 %d":          {1024, "%d", "1KB"},
		"1024 %.1f":        {1024, "%.1f", "1.0KB"},
		"1024 %s":          {1024, "%s", "1.024KB"},
		"3*MB+140*KB %f":   {3*int64(_MB) + 140*int64(_KB), "%f", "3.140000MB"},
		"3*MB+140*KB %d":   {3*int64(_MB) + 140*int64(_KB), "%d", "3MB"},
		"3*MB+140*KB %.1f": {3*int64(_MB) + 140*int64(_KB), "%.1f", "3.1MB"},
		"3*MB+140*KB %s":   {3*int64(_MB) + 140*int64(_KB), "%s", "3.14MB"},
		"2*GB %f":          {2 * int64(_GB), "%f", "2.000000GB"},
		"2*GB %d":          {2 * int64(_GB), "%d", "2GB"},
		"2*GB %.1f":        {2 * int64(_GB), "%.1f", "2.0GB"},
		"2*GB %s":          {2 * int64(_GB), "%s", "2GB"},
		"4*TB %f":          {4 * int64(_TB), "%f", "4.000000TB"},
		"4*TB %d":          {4 * int64(_TB), "%d", "4TB"},
		"4*TB %.1f":        {4 * int64(_TB), "%.1f", "4.0TB"},
		"4*TB %s":          {4 * int64(_TB), "%s", "4TB"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := fmt.Sprintf(tc.verb, SizeB1000(tc.value))
			if got != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, got)
			}
		})
	}
}
