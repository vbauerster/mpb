package mpb_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/vbauerster/mpb/v8"
)

type testWriter struct {
	io.Writer
	called bool
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.called = true
	return w.Writer.Write(p)
}

func TestProxyWriter(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	var buf bytes.Buffer
	tw := &testWriter{&buf, false}

	bar := p.New(int64(len(content)), mpb.NopStyle())

	_, err := io.Copy(bar.ProxyWriter(tw), strings.NewReader(content))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	p.Wait()

	if !tw.called {
		t.Error("Read not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}

type testWriteCloser struct {
	io.Writer
	called bool
}

func (w *testWriteCloser) Close() error {
	w.called = true
	return nil
}

func TestProxyWriteCloser(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	var buf bytes.Buffer
	tw := &testWriteCloser{&buf, false}

	bar := p.New(int64(len(content)), mpb.NopStyle())

	wc := bar.ProxyWriter(tw)
	_, err := io.Copy(wc, strings.NewReader(content))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}
	_ = wc.Close()

	p.Wait()

	if !tw.called {
		t.Error("Close not called")
	}
}

type testWriterReadFrom struct {
	io.Writer
	called bool
}

func (w *testWriterReadFrom) ReadFrom(r io.Reader) (n int64, err error) {
	w.called = true
	return w.Writer.(io.ReaderFrom).ReadFrom(r)
}

type dumbReader struct {
	r io.Reader
}

func (r dumbReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func TestProxyWriterReadFrom(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	var buf bytes.Buffer
	tw := &testWriterReadFrom{&buf, false}

	bar := p.New(int64(len(content)), mpb.NopStyle())

	// To trigger ReadFrom, WriteTo needs to be hidden, hence a dumb wrapper
	dr := dumbReader{strings.NewReader(content)}
	_, err := io.Copy(bar.ProxyWriter(tw), dr)
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	p.Wait()

	if !tw.called {
		t.Error("ReadFrom not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}
