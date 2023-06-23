package decor

import (
	"testing"
	"time"
)

func TestAverageSpeedSizeB1024(t *testing.T) {
	cases := []struct {
		name     string
		fmt      string
		unit     interface{}
		current  int64
		elapsed  time.Duration
		expected string
	}{
		{
			name:     "empty fmt",
			unit:     SizeB1024(0),
			fmt:      "",
			current:  0,
			elapsed:  time.Second,
			expected: "0 b/s",
		},
		{
			name:     "SizeB1024(0):%d:0b",
			unit:     SizeB1024(0),
			fmt:      "%d",
			current:  0,
			elapsed:  time.Second,
			expected: "0b/s",
		},
		{
			name:     "SizeB1024(0):%f:0b",
			unit:     SizeB1024(0),
			fmt:      "%f",
			current:  0,
			elapsed:  time.Second,
			expected: "0.000000b/s",
		},
		{
			name:     "SizeB1024(0):% .2f:0b",
			unit:     SizeB1024(0),
			fmt:      "% .2f",
			current:  0,
			elapsed:  time.Second,
			expected: "0.00 b/s",
		},
		{
			name:     "SizeB1024(0):%d:1b",
			unit:     SizeB1024(0),
			fmt:      "%d",
			current:  1,
			elapsed:  time.Second,
			expected: "1b/s",
		},
		{
			name:     "SizeB1024(0):% .2f:1b",
			unit:     SizeB1024(0),
			fmt:      "% .2f",
			current:  1,
			elapsed:  time.Second,
			expected: "1.00 b/s",
		},
		{
			name:     "SizeB1024(0):%d:KiB",
			unit:     SizeB1024(0),
			fmt:      "%d",
			current:  2 * int64(_iKiB),
			elapsed:  1 * time.Second,
			expected: "2KiB/s",
		},
		{
			name:     "SizeB1024(0):% .f:KiB",
			unit:     SizeB1024(0),
			fmt:      "% .2f",
			current:  2 * int64(_iKiB),
			elapsed:  1 * time.Second,
			expected: "2.00 KiB/s",
		},
		{
			name:     "SizeB1024(0):%d:MiB",
			unit:     SizeB1024(0),
			fmt:      "%d",
			current:  2 * int64(_iMiB),
			elapsed:  1 * time.Second,
			expected: "2MiB/s",
		},
		{
			name:     "SizeB1024(0):% .2f:MiB",
			unit:     SizeB1024(0),
			fmt:      "% .2f",
			current:  2 * int64(_iMiB),
			elapsed:  1 * time.Second,
			expected: "2.00 MiB/s",
		},
		{
			name:     "SizeB1024(0):%d:GiB",
			unit:     SizeB1024(0),
			fmt:      "%d",
			current:  2 * int64(_iGiB),
			elapsed:  1 * time.Second,
			expected: "2GiB/s",
		},
		{
			name:     "SizeB1024(0):% .2f:GiB",
			unit:     SizeB1024(0),
			fmt:      "% .2f",
			current:  2 * int64(_iGiB),
			elapsed:  1 * time.Second,
			expected: "2.00 GiB/s",
		},
		{
			name:     "SizeB1024(0):%d:TiB",
			unit:     SizeB1024(0),
			fmt:      "%d",
			current:  2 * int64(_iTiB),
			elapsed:  1 * time.Second,
			expected: "2TiB/s",
		},
		{
			name:     "SizeB1024(0):% .2f:TiB",
			unit:     SizeB1024(0),
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
			res, _ := decor.Decor(stat)
			if res != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, res)
			}
		})
	}
}

func TestAverageSpeedSizeB1000(t *testing.T) {
	cases := []struct {
		name     string
		fmt      string
		unit     interface{}
		current  int64
		elapsed  time.Duration
		expected string
	}{
		{
			name:     "empty fmt",
			unit:     SizeB1000(0),
			fmt:      "",
			current:  0,
			elapsed:  time.Second,
			expected: "0 b/s",
		},
		{
			name:     "SizeB1000(0):%d:0b",
			unit:     SizeB1000(0),
			fmt:      "%d",
			current:  0,
			elapsed:  time.Second,
			expected: "0b/s",
		},
		{
			name:     "SizeB1000(0):%f:0b",
			unit:     SizeB1000(0),
			fmt:      "%f",
			current:  0,
			elapsed:  time.Second,
			expected: "0.000000b/s",
		},
		{
			name:     "SizeB1000(0):% .2f:0b",
			unit:     SizeB1000(0),
			fmt:      "% .2f",
			current:  0,
			elapsed:  time.Second,
			expected: "0.00 b/s",
		},
		{
			name:     "SizeB1000(0):%d:1b",
			unit:     SizeB1000(0),
			fmt:      "%d",
			current:  1,
			elapsed:  time.Second,
			expected: "1b/s",
		},
		{
			name:     "SizeB1000(0):% .2f:1b",
			unit:     SizeB1000(0),
			fmt:      "% .2f",
			current:  1,
			elapsed:  time.Second,
			expected: "1.00 b/s",
		},
		{
			name:     "SizeB1000(0):%d:KB",
			unit:     SizeB1000(0),
			fmt:      "%d",
			current:  2 * int64(_KB),
			elapsed:  1 * time.Second,
			expected: "2KB/s",
		},
		{
			name:     "SizeB1000(0):% .f:KB",
			unit:     SizeB1000(0),
			fmt:      "% .2f",
			current:  2 * int64(_KB),
			elapsed:  1 * time.Second,
			expected: "2.00 KB/s",
		},
		{
			name:     "SizeB1000(0):%d:MB",
			unit:     SizeB1000(0),
			fmt:      "%d",
			current:  2 * int64(_MB),
			elapsed:  1 * time.Second,
			expected: "2MB/s",
		},
		{
			name:     "SizeB1000(0):% .2f:MB",
			unit:     SizeB1000(0),
			fmt:      "% .2f",
			current:  2 * int64(_MB),
			elapsed:  1 * time.Second,
			expected: "2.00 MB/s",
		},
		{
			name:     "SizeB1000(0):%d:GB",
			unit:     SizeB1000(0),
			fmt:      "%d",
			current:  2 * int64(_GB),
			elapsed:  1 * time.Second,
			expected: "2GB/s",
		},
		{
			name:     "SizeB1000(0):% .2f:GB",
			unit:     SizeB1000(0),
			fmt:      "% .2f",
			current:  2 * int64(_GB),
			elapsed:  1 * time.Second,
			expected: "2.00 GB/s",
		},
		{
			name:     "SizeB1000(0):%d:TB",
			unit:     SizeB1000(0),
			fmt:      "%d",
			current:  2 * int64(_TB),
			elapsed:  1 * time.Second,
			expected: "2TB/s",
		},
		{
			name:     "SizeB1000(0):% .2f:TB",
			unit:     SizeB1000(0),
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
			res, _ := decor.Decor(stat)
			if res != tc.expected {
				t.Fatalf("expected: %q, got: %q\n", tc.expected, res)
			}
		})
	}
}
