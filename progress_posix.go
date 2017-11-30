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

func (p *Progress) server(conf pConf) {
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)

	defer func() {
		if conf.shutdownNotifier != nil {
			close(conf.shutdownNotifier)
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
			op(&conf)
		case <-conf.ticker.C:
			if len(conf.bars) == 0 {
				runtime.Gosched()
				break
			}
			b0 := conf.bars[0]
			if numP == -1 {
				numP = b0.NumOfPrependers()
			}
			if numA == -1 {
				numA = b0.NumOfAppenders()
			}
			tw, _, _ := cwriter.TermSize()
			err := conf.writeAndFlush(tw, numP, numA)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case <-winch:
			tw, _, _ := cwriter.TermSize()
			err := conf.writeAndFlush(tw-tw/8, numP, numA)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if timer != nil && timer.Reset(resumeDelay) {
				break
			}
			conf.ticker.Stop()
			timer = time.NewTimer(resumeDelay)
			resumeTicker = timer.C
		case <-resumeTicker:
			conf.ticker = time.NewTicker(conf.rr)
			resumeTicker = nil
		case <-conf.cancel:
			conf.ticker.Stop()
			conf.cancel = nil
		case <-p.quit:
			if conf.cancel != nil {
				conf.ticker.Stop()
			}
			return
		}
	}
}
