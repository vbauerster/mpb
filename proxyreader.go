package mpb

import (
	"io"
	"time"
)

// proxyReader is io.Reader wrapper, for proxy read bytes
type proxyReader struct {
	r io.Reader
	b *Bar
}

func (s *proxyReader) Read(p []byte) (n int, err error) {
	start := time.Now()
	n, err = s.r.Read(p)
	if n > 0 {
		s.b.IncrBy(n, time.Since(start))
	}
	return
}

func (s *proxyReader) Close() error {
	if closer, ok := s.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
