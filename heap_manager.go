package mpb

import "container/heap"

type heapManager chan heapRequest

type heapCmd int

const (
	h_sync heapCmd = iota
	h_push
	h_iter
	h_iter2
	h_fix
	h_state
	h_end
)

type heapRequest struct {
	cmd  heapCmd
	data interface{}
}

type iter struct {
	drop <-chan struct{}
	dst  chan<- *Bar
}

type iter2 struct {
	drop <-chan struct{}
	dst  [2]chan<- *Bar
}

type pushData struct {
	bar  *Bar
	sync bool
}

type fixData struct {
	bar      *Bar
	priority int
	lazy     bool
}

func (m heapManager) run() {
	var bHeap barHeap
	var pMatrix, aMatrix map[int][]chan int

	var ln int
	var sync bool

	for req := range m {
		switch req.cmd {
		case h_push:
			data := req.data.(pushData)
			heap.Push(&bHeap, data.bar)
			sync = sync || data.sync
		case h_sync:
			if sync || ln != bHeap.Len() {
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
				ln = bHeap.Len()
			}
			drop := req.data.(<-chan struct{})
			syncWidth(pMatrix, drop)
			syncWidth(aMatrix, drop)
		case h_iter:
			data := req.data.(iter)
			rangeOverSlice(bHeap, data.dst, data.drop)
		case h_iter2:
			var dropped bool
			data := req.data.(iter2)
			for i, dst := range data.dst {
				if i == 0 {
					dropped = rangeOverSlice(bHeap, dst, data.drop)
				} else if !dropped {
					popOverHeap(&bHeap, dst, data.drop)
				}
			}
		case h_fix:
			data := req.data.(fixData)
			if data.bar.index < 0 {
				break
			}
			data.bar.priority = data.priority
			if !data.lazy {
				heap.Fix(&bHeap, data.bar.index)
			}
		case h_state:
			ch := req.data.(chan<- bool)
			ch <- sync || ln != bHeap.Len()
		case h_end:
			ch := req.data.(chan<- interface{})
			if ch != nil {
				go func() {
					ch <- []*Bar(bHeap)
				}()
			}
			return
		}
	}
}

func (m heapManager) sync(drop <-chan struct{}) {
	m <- heapRequest{cmd: h_sync, data: drop}
}

func (m heapManager) push(b *Bar, sync bool) {
	data := pushData{b, sync}
	req := heapRequest{cmd: h_push, data: data}
	select {
	case m <- req:
	default:
		go func() {
			m <- req
		}()
	}
}

func (m heapManager) iter(drop <-chan struct{}, iterRange chan<- *Bar) {
	data := iter{drop, iterRange}
	m <- heapRequest{cmd: h_iter, data: data}
}

func (m heapManager) iter2(drop <-chan struct{}, iterRange, iterPop chan<- *Bar) {
	data := iter2{drop, [2]chan<- *Bar{iterRange, iterPop}}
	m <- heapRequest{cmd: h_iter2, data: data}
}

func (m heapManager) fix(b *Bar, priority int, lazy bool) {
	data := fixData{b, priority, lazy}
	m <- heapRequest{cmd: h_fix, data: data}
}

func (m heapManager) state(ch chan<- bool) {
	m <- heapRequest{cmd: h_state, data: ch}
}

func (m heapManager) end(ch chan<- interface{}) {
	m <- heapRequest{cmd: h_end, data: ch}
}

func syncWidth(matrix map[int][]chan int, drop <-chan struct{}) {
	for _, column := range matrix {
		go maxWidthDistributor(column, drop)
	}
}

func maxWidthDistributor(column []chan int, drop <-chan struct{}) {
	var maxWidth int
	for _, ch := range column {
		select {
		case w := <-ch:
			if w > maxWidth {
				maxWidth = w
			}
		case <-drop:
			return
		}
	}
	for _, ch := range column {
		ch <- maxWidth
	}
}

// unordered iteration
func rangeOverSlice(s barHeap, dst chan<- *Bar, drop <-chan struct{}) (dropped bool) {
	defer close(dst)
	for _, b := range s {
		select {
		case dst <- b:
		case <-drop:
			return true
		}
	}
	return false
}

// ordered iteration
func popOverHeap(h heap.Interface, dst chan<- *Bar, drop <-chan struct{}) {
	defer close(dst)
	for h.Len() != 0 {
		bar := heap.Pop(h).(*Bar)
		select {
		case dst <- bar:
		case <-drop:
			heap.Push(h, bar)
			return
		}
	}
}
