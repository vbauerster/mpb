package mpb

import (
	"io"
	"time"
)

type proxyWriter struct {
	io.WriteCloser
	bar *Bar
}

func (x proxyWriter) Write(p []byte) (int, error) {
	n, err := x.WriteCloser.Write(p)
	x.bar.IncrBy(n)
	return n, err
}

type proxyReaderFrom struct {
	proxyWriter
}

func (x proxyReaderFrom) ReadFrom(r io.Reader) (int64, error) {
	n, err := x.WriteCloser.(io.ReaderFrom).ReadFrom(r)
	x.bar.IncrInt64(n)
	return n, err
}

type ewmaProxyWriter struct {
	proxyWriter
}

func (x ewmaProxyWriter) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := x.proxyWriter.Write(p)
	if n > 0 {
		x.bar.DecoratorEwmaUpdate(time.Since(start))
	}
	return n, err
}

type ewmaProxyReaderFrom struct {
	ewmaProxyWriter
}

func (x ewmaProxyReaderFrom) ReadFrom(r io.Reader) (int64, error) {
	start := time.Now()
	n, err := x.WriteCloser.(io.ReaderFrom).ReadFrom(r)
	if n > 0 {
		x.bar.DecoratorEwmaUpdate(time.Since(start))
	}
	return n, err
}

func newProxyWriter(w io.Writer, b *Bar, hasEwma bool) io.WriteCloser {
	pw := proxyWriter{toWriteCloser(w), b}
	if hasEwma {
		epw := ewmaProxyWriter{pw}
		if _, ok := w.(io.ReaderFrom); ok {
			return ewmaProxyReaderFrom{epw}
		}
		return epw
	}
	if _, ok := w.(io.ReaderFrom); ok {
		return proxyReaderFrom{pw}
	}
	return pw
}

func toWriteCloser(w io.Writer) io.WriteCloser {
	if wc, ok := w.(io.WriteCloser); ok {
		return wc
	}
	return toNopWriteCloser(w)
}

func toNopWriteCloser(w io.Writer) io.WriteCloser {
	if _, ok := w.(io.ReaderFrom); ok {
		return nopWriteCloserReaderFrom{w}
	}
	return nopWriteCloser{w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

type nopWriteCloserReaderFrom struct {
	io.Writer
}

func (nopWriteCloserReaderFrom) Close() error { return nil }

func (c nopWriteCloserReaderFrom) ReadFrom(r io.Reader) (int64, error) {
	return c.Writer.(io.ReaderFrom).ReadFrom(r)
}
