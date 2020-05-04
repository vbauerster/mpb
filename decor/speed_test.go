package decor

import (
	"testing"
	"time"
)

func TestSpeedKiBDecor(t *testing.T) {
	cases := []struct {
		name     string
		fmt      string
		unit     int
		current  int64
		elapsed  time.Duration
		expected string
	}{
		{
			name:     "empty fmt",
			unit:     UnitKiB,
			fmt:      "",
			current:  0,
			elapsed:  time.Second,
			expected: "0b/s",
		},
		{
			name:     "UnitKiB:%d:0b",
			unit:     UnitKiB,
			fmt:      "%d",
			current:  0,
			elapsed:  time.Second,
			expected: "0b/s",
		},
		{
			name:     "UnitKiB:% .2f:0b",
			unit:     UnitKiB,
			fmt:      "% .2f",
			current:  0,
			elapsed:  time.Second,
			expected: "0.00 b/s",
		},
		{
			name:     "UnitKiB:%d:1b",
			unit:     UnitKiB,
			fmt:      "%d",
			current:  1,
			elapsed:  time.Second,
			expected: "1b/s",
		},
		{
			name:     "UnitKiB:% .2f:1b",
			unit:     UnitKiB,
			fmt:      "% .2f",
			current:  1,
			elapsed:  time.Second,
			expected: "1.00 b/s",
		},
		{
			name:     "UnitKiB:%d:KiB",
			unit:     UnitKiB,
			fmt:      "%d",
			current:  2 * int64(_iKiB),
			elapsed:  1 * time.Second,
			expected: "2KiB/s",
		},
		{
			name:     "UnitKiB:% .f:KiB",
			unit:     UnitKiB,
			fmt:      "% .2f",
			current:  2 * int64(_iKiB),
			elapsed:  1 * time.Second,
			expected: "2.00 KiB/s",
		},
		{
			name:     "UnitKiB:%d:MiB",
			unit:     UnitKiB,
			fmt:      "%d",
			current:  2 * int64(_iMiB),
			elapsed:  1 * time.Second,
			expected: "2MiB/s",
		},
		{
			name:     "UnitKiB:% .2f:MiB",
			unit:     UnitKiB,
			fmt:      "% .2f",
			current:  2 * int64(_iMiB),
			elapsed:  1 * time.Second,
			expected: "2.00 MiB/s",
		},
		{
			name:     "UnitKiB:%d:GiB",
			unit:     UnitKiB,
			fmt:      "%d",
			current:  2 * int64(_iGiB),
			elapsed:  1 * time.Second,
			expected: "2GiB/s",
		},
		{
			name:     "UnitKiB:% .2f:GiB",
			unit:     UnitKiB,
			fmt:      "% .2f",
			current:  2 * int64(_iGiB),
			elapsed:  1 * time.Second,
			expected: "2.00 GiB/s",
		},
		{
			name:     "UnitKiB:%d:TiB",
			unit:     UnitKiB,
			fmt:      "%d",
			current:  2 * int64(_iTiB),
			elapsed:  1 * time.Second,
			expected: "2TiB/s",
		},
		{
			name:     "UnitKiB:% .2f:TiB",
			unit:     UnitKiB,
			fmt:      "% .2f",
			current:  2 * int64(_iTiB),
			elapsed:  1 * time.Second,
			expected: "2.00 TiB/s",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decor := NewAverageSpeed(tc.unit, tc.fmt, time.Now().Add(-tc.elapsed))
			stat := Statistics{
				Current: tc.current,
			}
			res := decor.Decor(stat)
			if res != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, res)
			}
		})
	}
}

func TestSpeedKBDecor(t *testing.T) {
	cases := []struct {
		name     string
		fmt      string
		unit     int
		current  int64
		elapsed  time.Duration
		expected string
	}{
		{
			name:     "empty fmt",
			unit:     UnitKB,
			fmt:      "",
			current:  0,
			elapsed:  time.Second,
			expected: "0b/s",
		},
		{
			name:     "UnitKB:%d:0b",
			unit:     UnitKB,
			fmt:      "%d",
			current:  0,
			elapsed:  time.Second,
			expected: "0b/s",
		},
		{
			name:     "UnitKB:% .2f:0b",
			unit:     UnitKB,
			fmt:      "% .2f",
			current:  0,
			elapsed:  time.Second,
			expected: "0.00 b/s",
		},
		{
			name:     "UnitKB:%d:1b",
			unit:     UnitKB,
			fmt:      "%d",
			current:  1,
			elapsed:  time.Second,
			expected: "1b/s",
		},
		{
			name:     "UnitKB:% .2f:1b",
			unit:     UnitKB,
			fmt:      "% .2f",
			current:  1,
			elapsed:  time.Second,
			expected: "1.00 b/s",
		},
		{
			name:     "UnitKB:%d:KB",
			unit:     UnitKB,
			fmt:      "%d",
			current:  2 * int64(_KB),
			elapsed:  1 * time.Second,
			expected: "2KB/s",
		},
		{
			name:     "UnitKB:% .f:KB",
			unit:     UnitKB,
			fmt:      "% .2f",
			current:  2 * int64(_KB),
			elapsed:  1 * time.Second,
			expected: "2.00 KB/s",
		},
		{
			name:     "UnitKB:%d:MB",
			unit:     UnitKB,
			fmt:      "%d",
			current:  2 * int64(_MB),
			elapsed:  1 * time.Second,
			expected: "2MB/s",
		},
		{
			name:     "UnitKB:% .2f:MB",
			unit:     UnitKB,
			fmt:      "% .2f",
			current:  2 * int64(_MB),
			elapsed:  1 * time.Second,
			expected: "2.00 MB/s",
		},
		{
			name:     "UnitKB:%d:GB",
			unit:     UnitKB,
			fmt:      "%d",
			current:  2 * int64(_GB),
			elapsed:  1 * time.Second,
			expected: "2GB/s",
		},
		{
			name:     "UnitKB:% .2f:GB",
			unit:     UnitKB,
			fmt:      "% .2f",
			current:  2 * int64(_GB),
			elapsed:  1 * time.Second,
			expected: "2.00 GB/s",
		},
		{
			name:     "UnitKB:%d:TB",
			unit:     UnitKB,
			fmt:      "%d",
			current:  2 * int64(_TB),
			elapsed:  1 * time.Second,
			expected: "2TB/s",
		},
		{
			name:     "UnitKB:% .2f:TB",
			unit:     UnitKB,
			fmt:      "% .2f",
			current:  2 * int64(_TB),
			elapsed:  1 * time.Second,
			expected: "2.00 TB/s",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decor := NewAverageSpeed(tc.unit, tc.fmt, time.Now().Add(-tc.elapsed))
			stat := Statistics{
				Current: tc.current,
			}
			res := decor.Decor(stat)
			if res != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, res)
			}
		})
	}
}
