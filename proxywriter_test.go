package mpb_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type testWriter struct {
	w      io.Writer
	called bool
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.called = true
	return w.w.Write(p)
}

func TestProxyWriter(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	tw := &testWriter{&buf, false}
	_, err := io.Copy(bar.ProxyWriter(tw), strings.NewReader(content))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	if !tw.called {
		t.Error("Read not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
	p.Wait()
}

func TestEwmaProxyWriter(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)),
		mpb.NopStyle(),
		mpb.AppendDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 30)),
	)

	var buf bytes.Buffer
	tw := &testWriter{&buf, false}
	_, err := io.Copy(bar.ProxyWriter(tw), strings.NewReader(content))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	if !tw.called {
		t.Error("Read not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
	p.Wait()
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
	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	tw := &testWriteCloser{&buf, false}
	wc := bar.ProxyWriter(tw)
	_, err := io.Copy(wc, strings.NewReader(content))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}
	_ = wc.Close()
	if !tw.called {
		t.Error("Close not called")
	}
	p.Wait()
}

func TestEwmaProxyWriteCloser(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)),
		mpb.NopStyle(),
		mpb.AppendDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 30)),
	)

	var buf bytes.Buffer
	tw := &testWriteCloser{&buf, false}
	wc := bar.ProxyWriter(tw)
	_, err := io.Copy(wc, strings.NewReader(content))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}
	_ = wc.Close()
	if !tw.called {
		t.Error("Close not called")
	}
	p.Wait()
}

type testWriterReadFrom struct {
	io.Writer
	called bool
}

func (w *testWriterReadFrom) Write(p []byte) (n int, err error) {
	return 0, errors.New("unexpected")
}

func (w *testWriterReadFrom) ReadFrom(r io.Reader) (n int64, err error) {
	w.called = true
	return w.Writer.(io.ReaderFrom).ReadFrom(r)
}

type wrapReader struct {
	r io.Reader
}

func (r wrapReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func TestProxyWriterReadFrom(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	tw := &testWriterReadFrom{&buf, false}
	// To trigger ReadFrom, WriteTo of strings.NewReader needs to be hidden
	_, err := io.Copy(bar.ProxyWriter(tw), wrapReader{strings.NewReader(content)})
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	if !tw.called {
		t.Error("ReadFrom not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
	p.Wait()
}

func TestEwmaProxyWriterReadFrom(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)),
		mpb.NopStyle(),
		mpb.AppendDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 30)),
	)

	var buf bytes.Buffer
	tw := &testWriterReadFrom{&buf, false}
	// To trigger ReadFrom, WriteTo of strings.NewReader needs to be hidden
	_, err := io.Copy(bar.ProxyWriter(tw), wrapReader{strings.NewReader(content)})
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	if !tw.called {
		t.Error("ReadFrom not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
	p.Wait()
}
