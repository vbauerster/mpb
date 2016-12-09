package uiprogress

import "bytes"

var (
	// Fill is the default character representing completed progress
	Fill byte = '='

	// Head is the default character that moves when progress is updated
	Head byte = '>'

	// Empty is the default character that represents the empty progress
	Empty byte = '-'

	// LeftEnd is the default character in the left most part of the progress indicator
	LeftEnd byte = '['

	// RightEnd is the default character in the right most part of the progress indicator
	RightEnd byte = ']'

	// Width is the default width of the progress bar
	Width = 70

	// ErrMaxCurrentReached is error when trying to set current value that exceeds the total value
	// ErrMaxCurrentReached = errors.New("errors: current value is greater total value")
)

// Bar represents a progress bar
type Bar struct {
	// Total of the total  for the progress bar
	Total int

	// LeftEnd is character in the left most part of the progress indicator. Defaults to '['
	LeftEnd byte

	// RightEnd is character in the right most part of the progress indicator. Defaults to ']'
	RightEnd byte

	// Fill is the character representing completed progress. Defaults to '='
	Fill byte

	// Head is the character that moves when progress is updated.  Defaults to '>'
	Head byte

	// Empty is the character that represents the empty progress. Default is '-'
	Empty byte

	// Width is the width of the progress bar
	Width int

	currentUpdateCh chan int

	redrawRequestCh chan redrawRequest

	// appendFuncs  []DecoratorFunc
	// prependFuncs []DecoratorFunc
}

// DecoratorFunc is a function that can be prepended and appended to the progress bar
type DecoratorFunc func(b *Bar) string

// NewBar returns a new progress bar
func NewBar(total int) *Bar {
	b := &Bar{
		Total:           total,
		Width:           Width,
		LeftEnd:         LeftEnd,
		RightEnd:        RightEnd,
		Head:            Head,
		Fill:            Fill,
		Empty:           Empty,
		currentUpdateCh: make(chan int),
		redrawRequestCh: make(chan redrawRequest),
	}
	go b.server()
	return b
}

type redrawRequest struct {
	bufch chan []byte
}

func (b *Bar) server() {
	var current int
	// blockStartTime := time.Now()
	// timePerItemEstimate time.Duration
	// remainingTime       time.Duration
	for {
		select {
		case n := <-b.currentUpdateCh:
			current += n
			// blockStartTime = time.Now()
		case r := <-b.redrawRequestCh:
			r.bufch <- b.buffer(current)
		}
	}
}

func (b *Bar) Update(n int) {
	b.currentUpdateCh <- n
}

func (b *Bar) buffer(current int) []byte {
	completedPercent := int(100 * float64(current) / float64(b.Total))
	completedWidth := completedPercent * b.Width / 100

	// add fill and empty bits
	var buf bytes.Buffer
	for i := 0; i < completedWidth; i++ {
		buf.WriteByte(b.Fill)
	}
	for i := 0; i < b.Width-completedWidth; i++ {
		buf.WriteByte(b.Empty)
	}

	// set head bit
	pb := buf.Bytes()
	if completedWidth > 0 && completedWidth < b.Width {
		pb[completedWidth-1] = b.Head
	}

	// set left and right ends bits
	pb[0], pb[len(pb)-1] = b.LeftEnd, b.RightEnd

	// render append functions to the right of the bar
	// for _, f := range b.appendFuncs {
	// 	pb = append(pb, ' ')
	// 	pb = append(pb, []byte(f(b))...)
	// }

	// render prepend functions to the left of the bar
	// for _, f := range b.prependFuncs {
	// 	args := []byte(f(b))
	// 	args = append(args, ' ')
	// 	pb = append(args, pb...)
	// }
	return pb
}

// String returns the string representation of the bar
func (b *Bar) String() string {
	bufch := make(chan []byte)
	b.redrawRequestCh <- redrawRequest{bufch}
	return string(<-bufch)
}

// CompletedPercent return the percent completed
// func (b *Bar) CompletedPercent() int {
// 	return int(100 * float64(b.current) / float64(b.Total))
// }

// AppendFunc runs the decorator function and renders the output on the right of the progress bar
// func (b *Bar) AppendFunc(f DecoratorFunc) *Bar {
// 	b.appendFuncs = append(b.appendFuncs, f)
// 	return b
// }

// PrependFunc runs decorator function and render the output left the progress bar
// func (b *Bar) PrependFunc(f DecoratorFunc) *Bar {
// 	b.prependFuncs = append(b.prependFuncs, f)
// 	return b
// }

// AppendCompleted appends the completion percent to the progress bar
// func (b *Bar) AppendCompleted() *Bar {
// 	b.AppendFunc(func(b *Bar) string {
// 		return b.CompletedPercentString()
// 	})
// 	return b
// }

// CompletedPercentString returns the formatted string representation of the completed percent
// func (b *Bar) CompletedPercentString() string {
// 	return fmt.Sprintf("%3d%%", b.CompletedPercent())
// }

// PrependElapsed prepends the time elapsed to the begining of the bar
// func (b *Bar) PrependElapsed() *Bar {
// 	b.PrependFunc(func(b *Bar) string {
// 		return strutil.PadLeft(b.TimeElapsedString(), 5, ' ')
// 	})
// 	return b
// }
