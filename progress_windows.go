// +build windows

package mpb

import (
	"fmt"
	"os"
	"runtime"

	"github.com/vbauerster/mpb/cwriter"
)

func (p *Progress) serve(s *pState) {
	var numP, numA int
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
				s.waitAll()
				if s.shutdownNotifier != nil {
					close(s.shutdownNotifier)
				}
				close(p.done)
				return
			}
		}
	}
}
