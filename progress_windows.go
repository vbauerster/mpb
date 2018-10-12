// +build windows

package mpb

func (p *Progress) serve(s *pState) {
	// common operation for ticker for manual refresh, returns
	// bool to indicate if we are done [false] or should continue [true]
	tickRefresh := func() bool {
		if s.zeroWait {
			s.ticker.Stop()
			if s.shutdownNotifier != nil {
				close(s.shutdownNotifier)
			}
			close(p.done)
			close(p.refresh)
			return false
		}
		tw, err := s.cw.GetWidth()
		if err != nil {
			tw = s.width
		}
		s.render(tw)
		return true
	}

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case <-s.ticker.C:
			if !tickRefresh() {
				return
			}
		case <-p.refresh:
			if !tickRefresh() {
				return
			}
		}
	}
}
