// +build windows

package mpb

import (
	"fmt"
	"os"
	"runtime"

	"github.com/vbauerster/mpb/cwriter"
)

func (p *Progress) serve(s *pState) {
	defer func() {
		if s.shutdownNotifier != nil {
			close(s.shutdownNotifier)
		}
		close(p.done)
	}()

	var numP, numA int

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case <-s.ticker.C:
			if s.bHeap.Len() == 0 {
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
