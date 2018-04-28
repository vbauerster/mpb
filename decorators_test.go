package mpb_test

import (
	"sync"
	"testing"
	"time"

	. "github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func TestStaticName(t *testing.T) {
	tests := []struct {
		fn   decor.DecoratorFunc
		want string
	}{
		{
			fn:   decor.StaticName("Test", 0, 0),
			want: "Test",
		},
		{
			fn:   decor.StaticName("Test", len("Test"), 0),
			want: "Test",
		},
		{
			fn:   decor.StaticName("Test", 10, 0),
			want: "      Test",
		},
		{
			fn:   decor.StaticName("Test", 10, decor.DidentRight),
			want: "Test      ",
		},
	}

	for _, test := range tests {
		got := test.fn(nil, nil, nil)
		if got != test.want {
			t.Errorf("Want: %q, Got: %q\n", test.want, got)
		}
	}
}

type step struct {
	stat *decor.Statistics
	dfn  decor.DecoratorFunc
	want string
}

func TestPercentageDwidthSync(t *testing.T) {

	testCases := [][]step{
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 8},
				decor.Percentage(0, decor.DwidthSync),
				"8 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DwidthSync),
				"9 %",
			},
		},
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DwidthSync),
				" 9 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 10},
				decor.Percentage(0, decor.DwidthSync),
				"10 %",
			},
		},
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DwidthSync),
				"  9 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 100},
				decor.Percentage(0, decor.DwidthSync),
				"100 %",
			},
		},
	}

	testDecoratorConcurrently(t, testCases)
}

func TestPercentageDwidthSyncDidentRight(t *testing.T) {

	testCases := [][]step{
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 8},
				decor.Percentage(0, decor.DwidthSync|decor.DidentRight),
				"8 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DwidthSync|decor.DidentRight),
				"9 %",
			},
		},
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DwidthSync|decor.DidentRight),
				"9 % ",
			},
			{
				&decor.Statistics{Total: 100, Current: 10},
				decor.Percentage(0, decor.DwidthSync|decor.DidentRight),
				"10 %",
			},
		},
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DwidthSync|decor.DidentRight),
				"9 %  ",
			},
			{
				&decor.Statistics{Total: 100, Current: 100},
				decor.Percentage(0, decor.DwidthSync|decor.DidentRight),
				"100 %",
			},
		},
	}

	testDecoratorConcurrently(t, testCases)
}

func TestPercentageDSyncSpace(t *testing.T) {

	testCases := [][]step{
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 8},
				decor.Percentage(0, decor.DSyncSpace),
				" 8 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DSyncSpace),
				" 9 %",
			},
		},
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DSyncSpace),
				"  9 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 10},
				decor.Percentage(0, decor.DSyncSpace),
				" 10 %",
			},
		},
		[]step{
			{
				&decor.Statistics{Total: 100, Current: 9},
				decor.Percentage(0, decor.DSyncSpace),
				"   9 %",
			},
			{
				&decor.Statistics{Total: 100, Current: 100},
				decor.Percentage(0, decor.DSyncSpace),
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

	numBars := len(testCases[0])
	var wg sync.WaitGroup
	for _, columnCase := range testCases {
		wg.Add(numBars)
		timeout := make(chan struct{})
		time.AfterFunc(100*time.Millisecond, func() {
			close(timeout)
		})
		ws := NewWidthSync(timeout, numBars, 1)
		res := make([]chan string, numBars)
		for i := 0; i < numBars; i++ {
			res[i] = make(chan string, 1)
			go func(s step, ch chan string) {
				defer wg.Done()
				ch <- s.dfn(s.stat, ws.Accumulator[0], ws.Distributor[0])
			}(columnCase[i], res[i])
		}
		wg.Wait()

		var i int
		for got := range fanIn(res...) {
			want := columnCase[i].want
			if got != want {
				t.Errorf("Want: %q, Got: %q\n", want, got)
			}
			i++
		}
	}
}

func fanIn(in ...chan string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, ich := range in {
			ch <- <-ich
		}
	}()
	return ch
}
