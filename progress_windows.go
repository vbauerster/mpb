// +build windows

package mpb

import "time"

func (p *Progress) serve(s *pState) {

	ticker := time.NewTicker(s.rr)
	refreshCh := ticker.C

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case <-refreshCh:
			if s.zeroWait {
				ticker.Stop()
				if s.shutdownNotifier != nil {
					close(s.shutdownNotifier)
				}
				close(p.done)
				return
			}
			tw, err := s.cw.GetWidth()
			if err != nil {
				tw = s.width
			}
			s.render(tw)
		}
	}
}
