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
}

func (x proxyWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := x.ReadCloser.(io.WriterTo).WriteTo(w)
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
}

func (x ewmaProxyWriterTo) WriteTo(w io.Writer) (int64, error) {
	start := time.Now()
	n, err := x.ReadCloser.(io.WriterTo).WriteTo(w)
	if n > 0 {
		x.bar.DecoratorEwmaUpdate(time.Since(start))
	}
	return n, err
}

func newProxyReader(r io.Reader, b *Bar, hasEwma bool) io.ReadCloser {
	pr := proxyReader{toReadCloser(r), b}
	if hasEwma {
		epr := ewmaProxyReader{pr}
		if _, ok := r.(io.WriterTo); ok {
			return ewmaProxyWriterTo{epr}
		}
		return epr
	}
	if _, ok := r.(io.WriterTo); ok {
		return proxyWriterTo{pr}
	}
	return pr
}

func toReadCloser(r io.Reader) io.ReadCloser {
	if rc, ok := r.(io.ReadCloser); ok {
		return rc
	}
	return toNopCloser(r)
}

func toNopCloser(r io.Reader) io.ReadCloser {
	if _, ok := r.(io.WriterTo); ok {
		return nopCloserWriterTo{r}
	}
	return nopCloser{r}
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type nopCloserWriterTo struct {
	io.Reader
}

func (nopCloserWriterTo) Close() error { return nil }

func (c nopCloserWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	return c.Reader.(io.WriterTo).WriteTo(w)
}
