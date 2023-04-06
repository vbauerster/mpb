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
		"verb %d":    {12345678, "%d", "12MiB"},
		"verb %s":    {12345678, "%s", "12MiB"},
		"verb %f":    {12345678, "%f", "11.773756MiB"},
		"verb %.6f":  {12345678, "%.6f", "11.773756MiB"},
		"verb %.0f":  {12345678, "%.0f", "12MiB"},
		"verb %.1f":  {12345678, "%.1f", "11.8MiB"},
		"verb %.2f":  {12345678, "%.2f", "11.77MiB"},
		"verb %.3f":  {12345678, "%.3f", "11.774MiB"},
		"verb % d":   {12345678, "% d", "12 MiB"},
		"verb % s":   {12345678, "% s", "12 MiB"},
		"verb % f":   {12345678, "% f", "11.773756 MiB"},
		"verb % .6f": {12345678, "% .6f", "11.773756 MiB"},
		"verb % .0f": {12345678, "% .0f", "12 MiB"},
		"verb % .1f": {12345678, "% .1f", "11.8 MiB"},
		"verb % .2f": {12345678, "% .2f", "11.77 MiB"},
		"verb % .3f": {12345678, "% .3f", "11.774 MiB"},

		"1000 %d":   {1000, "%d", "1000b"},
		"1000 %s":   {1000, "%s", "1000b"},
		"1000 %f":   {1000, "%f", "1000.000000b"},
		"1000 %.6f": {1000, "%.6f", "1000.000000b"},
		"1000 %.0f": {1000, "%.0f", "1000b"},
		"1000 %.1f": {1000, "%.1f", "1000.0b"},
		"1000 %.2f": {1000, "%.2f", "1000.00b"},
		"1000 %.3f": {1000, "%.3f", "1000.000b"},
		"1024 %d":   {1024, "%d", "1KiB"},
		"1024 %s":   {1024, "%s", "1KiB"},
		"1024 %f":   {1024, "%f", "1.000000KiB"},
		"1024 %.6f": {1024, "%.6f", "1.000000KiB"},
		"1024 %.0f": {1024, "%.0f", "1KiB"},
		"1024 %.1f": {1024, "%.1f", "1.0KiB"},
		"1024 %.2f": {1024, "%.2f", "1.00KiB"},
		"1024 %.3f": {1024, "%.3f", "1.000KiB"},

		"3*MiB+100KiB %d":   {3*int64(_iMiB) + 100*int64(_iKiB), "%d", "3MiB"},
		"3*MiB+100KiB %s":   {3*int64(_iMiB) + 100*int64(_iKiB), "%s", "3MiB"},
		"3*MiB+100KiB %f":   {3*int64(_iMiB) + 100*int64(_iKiB), "%f", "3.097656MiB"},
		"3*MiB+100KiB %.6f": {3*int64(_iMiB) + 100*int64(_iKiB), "%.6f", "3.097656MiB"},
		"3*MiB+100KiB %.0f": {3*int64(_iMiB) + 100*int64(_iKiB), "%.0f", "3MiB"},
		"3*MiB+100KiB %.1f": {3*int64(_iMiB) + 100*int64(_iKiB), "%.1f", "3.1MiB"},
		"3*MiB+100KiB %.2f": {3*int64(_iMiB) + 100*int64(_iKiB), "%.2f", "3.10MiB"},
		"3*MiB+100KiB %.3f": {3*int64(_iMiB) + 100*int64(_iKiB), "%.3f", "3.098MiB"},

		"2*GiB %d":   {2 * int64(_iGiB), "%d", "2GiB"},
		"2*GiB %s":   {2 * int64(_iGiB), "%s", "2GiB"},
		"2*GiB %f":   {2 * int64(_iGiB), "%f", "2.000000GiB"},
		"2*GiB %.6f": {2 * int64(_iGiB), "%.6f", "2.000000GiB"},
		"2*GiB %.0f": {2 * int64(_iGiB), "%.0f", "2GiB"},
		"2*GiB %.1f": {2 * int64(_iGiB), "%.1f", "2.0GiB"},
		"2*GiB %.2f": {2 * int64(_iGiB), "%.2f", "2.00GiB"},
		"2*GiB %.3f": {2 * int64(_iGiB), "%.3f", "2.000GiB"},

		"4*TiB %d":   {4 * int64(_iTiB), "%d", "4TiB"},
		"4*TiB %s":   {4 * int64(_iTiB), "%s", "4TiB"},
		"4*TiB %f":   {4 * int64(_iTiB), "%f", "4.000000TiB"},
		"4*TiB %.6f": {4 * int64(_iTiB), "%.6f", "4.000000TiB"},
		"4*TiB %.0f": {4 * int64(_iTiB), "%.0f", "4TiB"},
		"4*TiB %.1f": {4 * int64(_iTiB), "%.1f", "4.0TiB"},
		"4*TiB %.2f": {4 * int64(_iTiB), "%.2f", "4.00TiB"},
		"4*TiB %.3f": {4 * int64(_iTiB), "%.3f", "4.000TiB"},
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
		"verb %d":    {12345678, "%d", "12MB"},
		"verb %s":    {12345678, "%s", "12MB"},
		"verb %f":    {12345678, "%f", "12.345678MB"},
		"verb %.6f":  {12345678, "%.6f", "12.345678MB"},
		"verb %.0f":  {12345678, "%.0f", "12MB"},
		"verb %.1f":  {12345678, "%.1f", "12.3MB"},
		"verb %.2f":  {12345678, "%.2f", "12.35MB"},
		"verb %.3f":  {12345678, "%.3f", "12.346MB"},
		"verb % d":   {12345678, "% d", "12 MB"},
		"verb % s":   {12345678, "% s", "12 MB"},
		"verb % f":   {12345678, "% f", "12.345678 MB"},
		"verb % .6f": {12345678, "% .6f", "12.345678 MB"},
		"verb % .0f": {12345678, "% .0f", "12 MB"},
		"verb % .1f": {12345678, "% .1f", "12.3 MB"},
		"verb % .2f": {12345678, "% .2f", "12.35 MB"},
		"verb % .3f": {12345678, "% .3f", "12.346 MB"},

		"1000 %d":   {1000, "%d", "1KB"},
		"1000 %s":   {1000, "%s", "1KB"},
		"1000 %f":   {1000, "%f", "1.000000KB"},
		"1000 %.6f": {1000, "%.6f", "1.000000KB"},
		"1000 %.0f": {1000, "%.0f", "1KB"},
		"1000 %.1f": {1000, "%.1f", "1.0KB"},
		"1000 %.2f": {1000, "%.2f", "1.00KB"},
		"1000 %.3f": {1000, "%.3f", "1.000KB"},
		"1024 %d":   {1024, "%d", "1KB"},
		"1024 %s":   {1024, "%s", "1KB"},
		"1024 %f":   {1024, "%f", "1.024000KB"},
		"1024 %.6f": {1024, "%.6f", "1.024000KB"},
		"1024 %.0f": {1024, "%.0f", "1KB"},
		"1024 %.1f": {1024, "%.1f", "1.0KB"},
		"1024 %.2f": {1024, "%.2f", "1.02KB"},
		"1024 %.3f": {1024, "%.3f", "1.024KB"},

		"3*MB+100*KB %d":   {3*int64(_MB) + 100*int64(_KB), "%d", "3MB"},
		"3*MB+100*KB %s":   {3*int64(_MB) + 100*int64(_KB), "%s", "3MB"},
		"3*MB+100*KB %f":   {3*int64(_MB) + 100*int64(_KB), "%f", "3.100000MB"},
		"3*MB+100*KB %.6f": {3*int64(_MB) + 100*int64(_KB), "%.6f", "3.100000MB"},
		"3*MB+100*KB %.0f": {3*int64(_MB) + 100*int64(_KB), "%.0f", "3MB"},
		"3*MB+100*KB %.1f": {3*int64(_MB) + 100*int64(_KB), "%.1f", "3.1MB"},
		"3*MB+100*KB %.2f": {3*int64(_MB) + 100*int64(_KB), "%.2f", "3.10MB"},
		"3*MB+100*KB %.3f": {3*int64(_MB) + 100*int64(_KB), "%.3f", "3.100MB"},

		"2*GB %d":   {2 * int64(_GB), "%d", "2GB"},
		"2*GB %s":   {2 * int64(_GB), "%s", "2GB"},
		"2*GB %f":   {2 * int64(_GB), "%f", "2.000000GB"},
		"2*GB %.6f": {2 * int64(_GB), "%.6f", "2.000000GB"},
		"2*GB %.0f": {2 * int64(_GB), "%.0f", "2GB"},
		"2*GB %.1f": {2 * int64(_GB), "%.1f", "2.0GB"},
		"2*GB %.2f": {2 * int64(_GB), "%.2f", "2.00GB"},
		"2*GB %.3f": {2 * int64(_GB), "%.3f", "2.000GB"},

		"4*TB %d":   {4 * int64(_TB), "%d", "4TB"},
		"4*TB %s":   {4 * int64(_TB), "%s", "4TB"},
		"4*TB %f":   {4 * int64(_TB), "%f", "4.000000TB"},
		"4*TB %.6f": {4 * int64(_TB), "%.6f", "4.000000TB"},
		"4*TB %.0f": {4 * int64(_TB), "%.0f", "4TB"},
		"4*TB %.1f": {4 * int64(_TB), "%.1f", "4.0TB"},
		"4*TB %.2f": {4 * int64(_TB), "%.2f", "4.00TB"},
		"4*TB %.3f": {4 * int64(_TB), "%.3f", "4.000TB"},
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
