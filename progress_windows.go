// +build windows

package mpb

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
			tw, err := s.cw.GetWidth()
			if err != nil {
				tw = s.width
			}
			s.render(tw, numP, numA)
		}
	}
}
