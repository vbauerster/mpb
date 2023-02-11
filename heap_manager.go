package mpb

import (
	"container/heap"
)

type heapManager chan heapRequest

type heapCmd int

const (
	h_sync heapCmd = iota
	h_push
	h_iter
	h_drain
	h_fix
	h_end
)

type heapRequest struct {
	cmd  heapCmd
	data interface{}
}

type iterData struct {
	iter chan *Bar
	drop chan struct{}
}

type pushData struct {
	bar  *Bar
	sync bool
}

func (m heapManager) run() {
	var bHeap priorityQueue
	var pMatrix map[int][]chan int
	var aMatrix map[int][]chan int

	var l int
	var sync bool

	for req := range m {
		switch req.cmd {
		case h_push:
			data := req.data.(pushData)
			heap.Push(&bHeap, data.bar)
			if !sync {
				sync = data.sync
			}
		case h_sync:
			if sync || l != bHeap.Len() {
				pMatrix = make(map[int][]chan int)
				aMatrix = make(map[int][]chan int)
				for _, b := range bHeap {
					table := b.wSyncTable()
					for i, ch := range table[0] {
						pMatrix[i] = append(pMatrix[i], ch)
					}
					for i, ch := range table[1] {
						aMatrix[i] = append(aMatrix[i], ch)
					}
				}
				sync = false
				l = bHeap.Len()
			}
			syncWidth(pMatrix)
			syncWidth(aMatrix)
		case h_iter:
			data := req.data.(iterData)
			for _, b := range bHeap {
				select {
				case data.iter <- b:
				case <-data.drop:
				}
			}
			close(data.iter)
		case h_drain:
			data := req.data.(iterData)
			for bHeap.Len() != 0 {
				select {
				case data.iter <- heap.Pop(&bHeap).(*Bar):
				case <-data.drop:
				}
			}
			close(data.iter)
		case h_fix:
			heap.Fix(&bHeap, req.data.(int))
		case h_end:
			ch := req.data.(chan<- interface{})
			if ch != nil {
				go func() {
					ch <- []*Bar(bHeap)
				}()
			}
			close(m)
		}
	}
}

func (m heapManager) sync() {
	m <- heapRequest{cmd: h_sync}
}

func (m heapManager) push(b *Bar, sync bool) {
	data := pushData{b, sync}
	m <- heapRequest{cmd: h_push, data: data}
}

func (m heapManager) iter(iter chan *Bar, drop chan struct{}) {
	data := iterData{iter, drop}
	m <- heapRequest{cmd: h_iter, data: data}
}

func (m heapManager) drain(iter chan *Bar, drop chan struct{}) {
	data := iterData{iter, drop}
	m <- heapRequest{cmd: h_drain, data: data}
}

func (m heapManager) fix(index int) {
	m <- heapRequest{cmd: h_push, data: index}
}

func (m heapManager) end(ch chan<- interface{}) {
	m <- heapRequest{cmd: h_end, data: ch}
}

func syncWidth(matrix map[int][]chan int) {
	for _, column := range matrix {
		go maxWidthDistributor(column)
	}
}

func maxWidthDistributor(column []chan int) {
	var maxWidth int
	for _, ch := range column {
		if w := <-ch; w > maxWidth {
			maxWidth = w
		}
	}
	for _, ch := range column {
		ch <- maxWidth
	}
}
