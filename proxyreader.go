package mpb

import (
	"io"
	"time"
)

type proxyReader struct {
	io.ReadCloser
	bar *Bar
}

func (x proxyReader) Read(p []byte) (int, error) {
	n, err := x.ReadCloser.Read(p)
	x.bar.IncrBy(n)
	return n, err
}

type proxyWriterTo struct {
	proxyReader
	wt io.WriterTo
}

func (x proxyWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := x.wt.WriteTo(w)
	x.bar.IncrInt64(n)
	return n, err
}

type ewmaProxyReader struct {
	proxyReader
}

func (x ewmaProxyReader) Read(p []byte) (int, error) {
	start := time.Now()
	n, err := x.proxyReader.Read(p)
	if n > 0 {
		x.bar.DecoratorEwmaUpdate(time.Since(start))
	}
	return n, err
}

type ewmaProxyWriterTo struct {
	ewmaProxyReader
	wt proxyWriterTo
}

func (x ewmaProxyWriterTo) WriteTo(w io.Writer) (int64, error) {
	start := time.Now()
	n, err := x.wt.WriteTo(w)
	if n > 0 {
		x.bar.DecoratorEwmaUpdate(time.Since(start))
	}
	return n, err
}

func (b *Bar) newProxyReader(r io.Reader, hasEwma bool) io.ReadCloser {
	pr := proxyReader{toReadCloser(r), b}
	if hasEwma {
		epr := ewmaProxyReader{pr}
		if wt, ok := r.(io.WriterTo); ok {
			pwt := proxyWriterTo{pr, wt}
			return ewmaProxyWriterTo{epr, pwt}
		}
		return epr
	}
	if wt, ok := r.(io.WriterTo); ok {
		return proxyWriterTo{pr, wt}
	}
	return pr
}

func toReadCloser(r io.Reader) io.ReadCloser {
	if rc, ok := r.(io.ReadCloser); ok {
		return rc
	}
	return io.NopCloser(r)
}
