package mpb_test

import (
	"testing"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func TestNameDecorator(t *testing.T) {
	tests := []struct {
		decorator decor.Decorator
		want      string
	}{
		{
			decorator: decor.Name("Test"),
			want:      "Test",
		},
		{
			decorator: decor.Name("Test", decor.WC{W: len("Test")}),
			want:      "Test",
		},
		{
			decorator: decor.Name("Test", decor.WC{W: 10}),
			want:      "      Test",
		},
		{
			decorator: decor.Name("Test", decor.WC{W: 10, C: decor.DindentRight}),
			want:      "Test      ",
		},
	}

	for _, test := range tests {
		got, _ := test.decorator.Decor(decor.Statistics{})
		if got != test.want {
			t.Errorf("Want: %q, Got: %q\n", test.want, got)
		}
	}
}

type step struct {
	stat      decor.Statistics
	decorator decor.Decorator
	want      string
}

func TestPercentageDwidthSync(t *testing.T) {

	testCases := [][]step{
		{
			{
				decor.Statistics{Total: 100, Current: 8},
				decor.Percentage(decor.WCSyncWidth),
				"8 %",
			},
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncWidth),
				"9 %",
			},
		},
		{
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncWidth),
				" 9 %",
			},
			{
				decor.Statistics{Total: 100, Current: 10},
				decor.Percentage(decor.WCSyncWidth),
				"10 %",
			},
		},
		{
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncWidth),
				"  9 %",
			},
			{
				decor.Statistics{Total: 100, Current: 100},
				decor.Percentage(decor.WCSyncWidth),
				"100 %",
			},
		},
	}

	testDecoratorConcurrently(t, testCases)
}

func TestPercentageDwidthSyncDindentRight(t *testing.T) {

	testCases := [][]step{
		{
			{
				decor.Statistics{Total: 100, Current: 8},
				decor.Percentage(decor.WCSyncWidthR),
				"8 %",
			},
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncWidthR),
				"9 %",
			},
		},
		{
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncWidthR),
				"9 % ",
			},
			{
				decor.Statistics{Total: 100, Current: 10},
				decor.Percentage(decor.WCSyncWidthR),
				"10 %",
			},
		},
		{
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncWidthR),
				"9 %  ",
			},
			{
				decor.Statistics{Total: 100, Current: 100},
				decor.Percentage(decor.WCSyncWidthR),
				"100 %",
			},
		},
	}

	testDecoratorConcurrently(t, testCases)
}

func TestPercentageDSyncSpace(t *testing.T) {

	testCases := [][]step{
		{
			{
				decor.Statistics{Total: 100, Current: 8},
				decor.Percentage(decor.WCSyncSpace),
				" 8 %",
			},
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncSpace),
				" 9 %",
			},
		},
		{
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncSpace),
				"  9 %",
			},
			{
				decor.Statistics{Total: 100, Current: 10},
				decor.Percentage(decor.WCSyncSpace),
				" 10 %",
			},
		},
		{
			{
				decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(decor.WCSyncSpace),
				"   9 %",
			},
			{
				decor.Statistics{Total: 100, Current: 100},
				decor.Percentage(decor.WCSyncSpace),
				" 100 %",
			},
		},
	}

	testDecoratorConcurrently(t, testCases)
}

func testDecoratorConcurrently(t *testing.T, testCases [][]step) {
	if len(testCases) == 0 {
		t.Fail()
	}

	for _, columnCase := range testCases {
		mpb.SyncWidth(toSyncMatrix(columnCase), nil)
		var results []chan string
		for _, step := range columnCase {
			step := step
			ch := make(chan string)
			go func() {
				str, _ := step.decorator.Decor(step.stat)
				ch <- str
			}()
			results = append(results, ch)
		}

		for i, ch := range results {
			res := <-ch
			want := columnCase[i].want
			if res != want {
				t.Errorf("Want: %q, Got: %q\n", want, res)
			}
		}
	}
}

func toSyncMatrix(ss []step) map[int][]chan int {
	var column []chan int
	for _, s := range ss {
		if ch, ok := s.decorator.Sync(); ok {
			column = append(column, ch)
		}
	}
	return map[int][]chan int{0: column}
}
