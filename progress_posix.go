// +build !windows

package mpb

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/vbauerster/mpb/cwriter"
)

func (p *Progress) server(s *pState) {
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)

	defer func() {
		if s.shutdownNotifier != nil {
			close(s.shutdownNotifier)
		}
		signal.Stop(winch)
		close(p.done)
	}()

	numP, numA := -1, -1

	var timer *time.Timer
	var resumeTicker <-chan time.Time
	resumeDelay := 300 * time.Millisecond

	for {
		select {
		case op := <-p.ops:
			op(s)
		case <-s.ticker.C:
			if len(s.bars) == 0 {
				runtime.Gosched()
				break
			}
			b0 := s.bars[0]
			if numP == -1 {
				numP = b0.NumOfPrependers()
			}
			if numA == -1 {
				numA = b0.NumOfAppenders()
			}
			tw, _, _ := cwriter.TermSize()
			err := s.writeAndFlush(tw, numP, numA)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case <-winch:
			tw, _, _ := cwriter.TermSize()
			err := s.writeAndFlush(tw-tw/8, numP, numA)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if timer != nil && timer.Reset(resumeDelay) {
				break
			}
			s.ticker.Stop()
			timer = time.NewTimer(resumeDelay)
			resumeTicker = timer.C
		case <-resumeTicker:
			s.ticker = time.NewTicker(s.rr)
			resumeTicker = nil
		case <-s.cancel:
			s.ticker.Stop()
			s.cancel = nil
		case <-p.quit:
			if s.cancel != nil {
				s.ticker.Stop()
			}
			return
		}
	}
}
