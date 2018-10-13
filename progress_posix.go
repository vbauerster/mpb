// +build !windows

package mpb

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (p *Progress) serve(s *pState) {

	var resumeTimer *time.Timer
	var resumeEvent <-chan time.Time
	winchIdleDur := s.rr * 2
	winch := make(chan os.Signal, 2)
	signal.Notify(winch, syscall.SIGWINCH)

	ticker := time.NewTicker(s.rr)
	refreshCh := ticker.C

	for {
		select {
		case op := <-p.operateState:
			op(s)
		case <-refreshCh:
			if s.zeroWait {
				ticker.Stop()
				signal.Stop(winch)
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
		case <-winch:
			tw, err := s.cw.GetWidth()
			if err != nil {
				tw = s.width
			}
			s.render(tw - tw/8)
			if resumeTimer != nil && resumeTimer.Reset(winchIdleDur) {
				break
			}
			ticker.Stop()
			resumeTimer = time.NewTimer(winchIdleDur)
			resumeEvent = resumeTimer.C
		case <-resumeEvent:
			ticker = time.NewTicker(s.rr)
			refreshCh = ticker.C
			resumeEvent = nil
			resumeTimer = nil
		}
	}
}
