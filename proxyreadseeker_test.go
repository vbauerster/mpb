package mpb_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func TestProxyReadSeeker(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	total := int64(len(content))
	bar := p.AddBar(total,
		mpb.AppendDecorators(decor.Percentage()),
	)

	rs, err := bar.ProxyReadSeeker(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 32)
	var read int64
	for {
		n, err := rs.Read(buf)
		read += int64(n)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected read error: %v", err)
		}
	}

	if read != total {
		t.Errorf("read %d bytes, want %d", read, total)
	}
	if bar.Current() != total {
		t.Errorf("bar.Current() = %d, want %d", bar.Current(), total)
	}
	p.Wait()
}

func TestProxyReadSeekerSeekRewind(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	total := int64(len(content))
	bar := p.AddBar(total)

	rs, err := bar.ProxyReadSeeker(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	// Read half
	half := total / 2
	_, _ = io.CopyN(io.Discard, rs, half)
	if bar.Current() != half {
		t.Errorf("after half read: bar.Current() = %d, want %d", bar.Current(), half)
	}

	// Seek back to start (simulates S3 retry)
	pos, err := rs.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatalf("seek error: %v", err)
	}
	if pos != 0 {
		t.Errorf("seek returned %d, want 0", pos)
	}
	if bar.Current() != 0 {
		t.Errorf("after seek to start: bar.Current() = %d, want 0", bar.Current())
	}

	// Read everything again
	_, _ = io.Copy(io.Discard, rs)
	if bar.Current() != total {
		t.Errorf("after full re-read: bar.Current() = %d, want %d", bar.Current(), total)
	}
	p.Wait()
}

func TestProxyReadSeekerSeekEnd(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	total := int64(len(content))
	bar := p.AddBar(total)

	rs, err := bar.ProxyReadSeeker(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	pos, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatalf("seek error: %v", err)
	}
	if pos != total {
		t.Errorf("seek to end returned %d, want %d", pos, total)
	}
	if bar.Current() != total {
		t.Errorf("after seek to end: bar.Current() = %d, want %d", bar.Current(), total)
	}
	bar.SetTotal(total, true)
	p.Wait()
}

func TestProxyReadSeekerClose(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	total := int64(len(content))
	bar := p.AddBar(total)

	rc := &readSeekNopCloser{ReadSeeker: strings.NewReader(content)}
	rs, err := bar.ProxyReadSeeker(rc)
	if err != nil {
		t.Fatal(err)
	}

	// Read everything so the bar completes naturally
	_, _ = io.Copy(io.Discard, rs)

	if err := rs.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	if !rc.closed {
		t.Error("expected underlying ReadCloser to be closed")
	}
	p.Wait()
}

func TestProxyReadSeekerNilPanics(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.AddBar(10)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ReadSeeker")
		}
		p.Shutdown()
	}()
	_, _ = bar.ProxyReadSeeker(nil)
}

func TestProxyReadSeekerDataIntegrity(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	total := int64(len(content))
	bar := p.AddBar(total)

	rs, err := bar.ProxyReadSeeker(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rs)

	if buf.String() != content {
		t.Error("proxied content does not match original")
	}
	p.Wait()
}

type readSeekNopCloser struct {
	io.ReadSeeker
	closed bool
}

func (r *readSeekNopCloser) Close() error {
	r.closed = true
	return nil
}
