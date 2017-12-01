// +build windows

package mpb

import (
	"fmt"
	"os"
	"runtime"

	"github.com/vbauerster/mpb/cwriter"
)

func (p *Progress) server(s *pState) {
	defer func() {
		if s.shutdownNotifier != nil {
			close(s.shutdownNotifier)
		}
		close(p.done)
	}()

	numP, numA := -1, -1

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
