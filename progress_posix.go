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

func (p *Progress) serve(s *pState) {
	winch := make(chan os.Signal, 2)
	signal.Notify(winch, syscall.SIGWINCH)

	var numP, numA int
	var timer *time.Timer
	var resumeTicker <-chan time.Time
	resumeDelay := 300 * time.Millisecond

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case <-s.ticker.C:
			if s.bHeap.Len() == 0 {
				if s.zeroWait {
					close(p.done)
					return
				}
				runtime.Gosched()
				break
			}
			if s.heapUpdated {
				numP = s.bHeap.maxNumP()
				numA = s.bHeap.maxNumA()
				s.heapUpdated = false
			}
			tw, _, _ := cwriter.TermSize()
			err := s.writeAndFlush(tw, numP, numA)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			var completed int
			for i := 0; i < s.bHeap.Len(); i++ {
				b := (*s.bHeap)[i]
				if b.completed {
					completed++
				}
			}
			if completed == s.bHeap.Len() {
				s.ticker.Stop()
				signal.Stop(winch)
				s.waitAll()
				if s.shutdownNotifier != nil {
					close(s.shutdownNotifier)
				}
				close(p.done)
				return
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
		}
	}
}
