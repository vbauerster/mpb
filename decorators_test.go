package mpb_test

import (
	"sync"
	"testing"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
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
			decorator: decor.Name("Test", decor.WC{W: 10, C: decor.DidentRight}),
			want:      "Test      ",
		},
	}

	for _, test := range tests {
		got := test.decorator.Decor(decor.Statistics{})
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

func TestPercentageDwidthSyncDidentRight(t *testing.T) {

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
		mpb.SyncWidth(toSyncMatrix(columnCase))
		numBars := len(columnCase)
		gott := make([]chan string, numBars)
		wg := new(sync.WaitGroup)
		wg.Add(numBars)
		for i, step := range columnCase {
			step := step
			ch := make(chan string, 1)
			go func() {
				defer wg.Done()
				ch <- step.decorator.Decor(step.stat)
			}()
			gott[i] = ch
		}
		wg.Wait()

		for i, ch := range gott {
			got := <-ch
			want := columnCase[i].want
			if got != want {
				t.Errorf("Want: %q, Got: %q\n", want, got)
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
