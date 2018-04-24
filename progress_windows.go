// +build windows

package mpb

import (
	"fmt"
	"time"

	"github.com/vbauerster/mpb/cwriter"
)

func (p *Progress) serve(s *pState) {
	var numP, numA int
	for {
		select {
		case op := <-p.operateState:
			op(s)
		case <-s.ticker.C:
			if s.zeroWait {
				s.ticker.Stop()
				if s.shutdownNotifier != nil {
					close(s.shutdownNotifier)
				}
				close(p.done)
				return
			}
			if s.heapUpdated {
				numP = s.bHeap.maxNumP()
				numA = s.bHeap.maxNumA()
				s.heapUpdated = false
			}
			tw, _, _ := cwriter.TermSize()
			err := s.writeAndFlush(tw, numP, numA)
			if err != nil {
				fmt.Fprintf(s.debugOut, "%s %s %v\n", "[mpb]", time.Now(), err)
			}
		}
	}
}
