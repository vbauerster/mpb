// +build windows

package mpb

import (
	"fmt"
	"os"
	"runtime"

	"github.com/vbauerster/mpb/cwriter"
)

func (p *Progress) server(conf pConf) {
	defer func() {
		if conf.shutdownNotifier != nil {
			close(conf.shutdownNotifier)
		}
		close(p.done)
	}()

	numP, numA := -1, -1

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
