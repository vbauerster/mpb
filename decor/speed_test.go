package decor

import (
	"fmt"
	"testing"
)

func TestSpeedKiB(t *testing.T) {
	cases := map[string]struct {
		value          int64
		verb, expected string
	}{
		"verb %f":   {12345678, "%f", "11.773756MiB/s"},
		"verb %.0f": {12345678, "%.0f", "12MiB/s"},
		"verb %.1f": {12345678, "%.1f", "11.8MiB/s"},
		"verb %.2f": {12345678, "%.2f", "11.77MiB/s"},
		"verb %.3f": {12345678, "%.3f", "11.774MiB/s"},

		"verb % f":   {12345678, "% f", "11.773756 MiB/s"},
		"verb % .0f": {12345678, "% .0f", "12 MiB/s"},
		"verb % .1f": {12345678, "% .1f", "11.8 MiB/s"},
		"verb % .2f": {12345678, "% .2f", "11.77 MiB/s"},
		"verb % .3f": {12345678, "% .3f", "11.774 MiB/s"},

		"verb %10.f":  {12345678, "%10.f", "   12MiB/s"},
		"verb %10.0f": {12345678, "%10.0f", "   12MiB/s"},
		"verb %10.1f": {12345678, "%10.1f", " 11.8MiB/s"},
		"verb %10.2f": {12345678, "%10.2f", "11.77MiB/s"},
		"verb %10.3f": {12345678, "%10.3f", "11.774MiB/s"},

		"verb % 10.f":  {12345678, "% 10.f", "  12 MiB/s"},
		"verb % 10.0f": {12345678, "% 10.0f", "  12 MiB/s"},
		"verb % 10.1f": {12345678, "% 10.1f", "11.8 MiB/s"},

		"verb %-10.f":  {12345678, "%-10.f", "12MiB/s   "},
		"verb %-10.0f": {12345678, "%-10.0f", "12MiB/s   "},
		"verb %-10.1f": {12345678, "%-10.1f", "11.8MiB/s "},
		"verb %-10.2f": {12345678, "%10.2f", "11.77MiB/s"},
		"verb %-10.3f": {12345678, "%10.3f", "11.774MiB/s"},

		"verb % -10.f":  {12345678, "% -10.f", "12 MiB/s  "},
		"verb % -10.0f": {12345678, "% -10.0f", "12 MiB/s  "},
		"verb % -10.1f": {12345678, "% -10.1f", "11.8 MiB/s"},

		"1000 %f":               {1000, "%f", "1000b/s"},
		"1000 %d":               {1000, "%d", "1000b/s"},
		"1000 %s":               {1000, "%s", "1000b/s"},
		"1024 %f":               {1024, "%f", "1.000000KiB/s"},
		"1024 %d":               {1024, "%d", "1KiB/s"},
		"1024 %.1f":             {1024, "%.1f", "1.0KiB/s"},
		"1024 %s":               {1024, "%s", "1.0KiB/s"},
		"3*MiB/s+140KiB/s %f":   {3*MiB + 140*KiB, "%f", "3.136719MiB/s"},
		"3*MiB/s+140KiB/s %d":   {3*MiB + 140*KiB, "%d", "3MiB/s"},
		"3*MiB/s+140KiB/s %.1f": {3*MiB + 140*KiB, "%.1f", "3.1MiB/s"},
		"3*MiB/s+140KiB/s %s":   {3*MiB + 140*KiB, "%s", "3.1MiB/s"},
		"2*GiB/s %f":            {2 * GiB, "%f", "2.000000GiB/s"},
		"2*GiB/s %d":            {2 * GiB, "%d", "2GiB/s"},
		"2*GiB/s %.1f":          {2 * GiB, "%.1f", "2.0GiB/s"},
		"2*GiB/s %s":            {2 * GiB, "%s", "2.0GiB/s"},
		"4*TiB/s %f":            {4 * TiB, "%f", "4.000000TiB/s"},
		"4*TiB/s %d":            {4 * TiB, "%d", "4TiB/s"},
		"4*TiB/s %.1f":          {4 * TiB, "%.1f", "4.0TiB/s"},
		"4*TiB/s %s":            {4 * TiB, "%s", "4.0TiB/s"},
	}
	for k, tc := range cases {
		got := fmt.Sprintf(tc.verb, SpeedKiB(tc.value))
		if got != tc.expected {
			t.Errorf("%s: Expected: %q, got: %q\n", k, tc.expected, got)
		}
	}
}

func TestSpeedKB(t *testing.T) {
	cases := map[string]struct {
		value          int64
		verb, expected string
	}{
		"verb %f":   {12345678, "%f", "12.345678MB/s"},
		"verb %.0f": {12345678, "%.0f", "12MB/s"},
		"verb %.1f": {12345678, "%.1f", "12.3MB/s"},
		"verb %.2f": {12345678, "%.2f", "12.35MB/s"},
		"verb %.3f": {12345678, "%.3f", "12.346MB/s"},

		"verb % f":   {12345678, "% f", "12.345678 MB/s"},
		"verb % .0f": {12345678, "% .0f", "12 MB/s"},
		"verb % .1f": {12345678, "% .1f", "12.3 MB/s"},
		"verb % .2f": {12345678, "% .2f", "12.35 MB/s"},
		"verb % .3f": {12345678, "% .3f", "12.346 MB/s"},

		"verb %10.f":  {12345678, "%10.f", "    12MB/s"},
		"verb %10.0f": {12345678, "%10.0f", "    12MB/s"},
		"verb %10.1f": {12345678, "%10.1f", "  12.3MB/s"},
		"verb %10.2f": {12345678, "%10.2f", " 12.35MB/s"},
		"verb %10.3f": {12345678, "%10.3f", "12.346MB/s"},

		"verb % 10.f":  {12345678, "% 10.f", "   12 MB/s"},
		"verb % 10.0f": {12345678, "% 10.0f", "   12 MB/s"},
		"verb % 10.1f": {12345678, "% 10.1f", " 12.3 MB/s"},

		"verb %-10.f":  {12345678, "%-10.f", "12MB/s    "},
		"verb %-10.0f": {12345678, "%-10.0f", "12MB/s    "},
		"verb %-10.1f": {12345678, "%-10.1f", "12.3MB/s  "},
		"verb %-10.2f": {12345678, "%10.2f", " 12.35MB/s"},
		"verb %-10.3f": {12345678, "%10.3f", "12.346MB/s"},

		"verb % -10.f":  {12345678, "% -10.f", "12 MB/s   "},
		"verb % -10.0f": {12345678, "% -10.0f", "12 MB/s   "},
		"verb % -10.1f": {12345678, "% -10.1f", "12.3 MB/s "},

		"1000 %f":              {1000, "%f", "1.000000kB/s"},
		"1000 %d":              {1000, "%d", "1kB/s"},
		"1000 %s":              {1000, "%s", "1.0kB/s"},
		"1024 %f":              {1024, "%f", "1.024000kB/s"},
		"1024 %d":              {1024, "%d", "1kB/s"},
		"1024 %.1f":            {1024, "%.1f", "1.0kB/s"},
		"1024 %s":              {1024, "%s", "1.0kB/s"},
		"3*MB/s+140*KB/s %f":   {3*MB + 140*KB, "%f", "3.140000MB/s"},
		"3*MB/s+140*KB/s %d":   {3*MB + 140*KB, "%d", "3MB/s"},
		"3*MB/s+140*KB/s %.1f": {3*MB + 140*KB, "%.1f", "3.1MB/s"},
		"3*MB/s+140*KB/s %s":   {3*MB + 140*KB, "%s", "3.1MB/s"},
		"2*GB/s %f":            {2 * GB, "%f", "2.000000GB/s"},
		"2*GB/s %d":            {2 * GB, "%d", "2GB/s"},
		"2*GB/s %.1f":          {2 * GB, "%.1f", "2.0GB/s"},
		"2*GB/s %s":            {2 * GB, "%s", "2.0GB/s"},
		"4*TB/s %f":            {4 * TB, "%f", "4.000000TB/s"},
		"4*TB/s %d":            {4 * TB, "%d", "4TB/s"},
		"4*TB/s %.1f":          {4 * TB, "%.1f", "4.0TB/s"},
		"4*TB/s %s":            {4 * TB, "%s", "4.0TB/s"},
	}
	for k, tc := range cases {
		got := fmt.Sprintf(tc.verb, SpeedKB(tc.value))
		if got != tc.expected {
			t.Errorf("%s: Expected: %q, got: %q\n", k, tc.expected, got)
		}
	}
}
