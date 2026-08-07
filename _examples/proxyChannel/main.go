package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// chunk represents a piece of data flowing through a pipeline.
// It implements mpb.Sizer so ProxyChannel tracks bytes rather than items.
type chunk struct {
	id   int
	data []byte
}

func (c chunk) Size() int64 { return int64(len(c.data)) }

// job is a plain work item with no size. ProxyChannel counts each as 1.
type job struct{ id int }

func main() {
	const (
		numChunks    = 50
		chunkMaxSize = 128 * 1024 // 128 KiB
		numJobs      = 30
		updateRate   = 500 * time.Millisecond
	)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ── bar 1: chunks with byte-level tracking (Sizer) + EWMA speed/ETA ──────

	// Build chunk sizes upfront so we know the exact total for the bar.
	chunkSizes := make([]int, numChunks)
	var totalBytes int64
	for i := range chunkSizes {
		chunkSizes[i] = rng.Intn(chunkMaxSize) + 1
		totalBytes += int64(chunkSizes[i])
	}

	chunkCh := make(chan chunk, 8)
	go func() {
		localRng := rand.New(rand.NewSource(time.Now().UnixNano() + 1))
		for i, sz := range chunkSizes {
			chunkCh <- chunk{id: i, data: make([]byte, sz)}
			time.Sleep(time.Duration(localRng.Intn(40)+10) * time.Millisecond)
		}
		close(chunkCh)
	}()

	// ── bar 2: plain job items (no Sizer, each counts as 1) ──────────────────

	jobCh := make(chan job, 4)
	go func() {
		localRng := rand.New(rand.NewSource(time.Now().UnixNano() + 2))
		for i := range numJobs {
			jobCh <- job{id: i}
			time.Sleep(time.Duration(localRng.Intn(80)+20) * time.Millisecond)
		}
		close(jobCh)
	}()

	// ── progress container ────────────────────────────────────────────────────

	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(72))

	// Byte-counting bar: uses Sizer path of ProxyChannel + EWMA decorators.
	byteBar := p.New(totalBytes,
		mpb.BarStyle(),
		mpb.PrependDecorators(
			decor.Name("chunks ", decor.WC{C: decor.DindentRight}),
		),
		mpb.AppendDecorators(

			decor.Counters(decor.SizeB1024(0), "% .1f / % .1f", decor.WCSyncSpaceR),
			decor.OnComplete(
				decor.EwmaSpeed(decor.SizeB1024(0), "% .1f", 30, decor.WCSyncSpaceR), "",
			),
			decor.OnComplete(
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
			),
		),
	)

	// Item-counting bar: uses plain increment path of ProxyChannel.
	jobBar := p.New(int64(numJobs),
		mpb.BarStyle(),
		mpb.PrependDecorators(
			decor.Name("jobs   ", decor.WC{C: decor.DindentRight}),
		),
		mpb.AppendDecorators(
			decor.Counters(0, "%d / %d", decor.WCSyncSpaceR),
			decor.Percentage(decor.WCSyncSpaceR),
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WCSyncWidth), "done",
			),
		),
	)

	// ── consumers ─────────────────────────────────────────────────────────────

	wg.Add(2)

	// Consume byte chunks through the proxy; bar updates are batched by updateRate.
	go func() {
		defer wg.Done()
		out := mpb.ProxyChannel(byteBar, chunkCh, updateRate)
		for v := range out {
			c := v.(chunk)
			_ = c // real code would process c.data here
		}
	}()

	// Consume jobs through the proxy; each job counts as 1 (no Sizer).
	go func() {
		defer wg.Done()
		out := mpb.ProxyChannel(jobBar, jobCh, updateRate)
		for v := range out {
			j := v.(job)
			_ = j // real code would handle the job here
		}
	}()

	p.Wait()

	fmt.Println("all done.")
}
