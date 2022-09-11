package mpb_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/vbauerster/mpb/v7"
)

const content = `Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do
		eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim
		veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea
		commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit
		esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat
		cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id
		est laborum.`

type testReader struct {
	io.Reader
	called bool
}

func (r *testReader) Read(p []byte) (n int, err error) {
	r.called = true
	return r.Reader.Read(p)
}

func TestProxyReader(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	reader := &testReader{strings.NewReader(content), false}

	bar := p.AddBar(int64(len(content)))

	var buf bytes.Buffer
	_, err := io.Copy(&buf, bar.ProxyReader(reader))
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if !reader.called {
		t.Error("Read not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}

type testReadCloser struct {
	io.Reader
	called bool
}

func (r *testReadCloser) Close() error {
	r.called = true
	return nil
}

func TestProxyReadCloser(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	reader := &testReadCloser{strings.NewReader(content), false}

	bar := p.AddBar(int64(len(content)))

	rc := bar.ProxyReader(reader)
	_, _ = io.Copy(io.Discard, rc)
	_ = rc.Close()

	p.Wait()

	if !reader.called {
		t.Error("Close not called")
	}
}

type testWriterTo struct {
	io.Reader
	called bool
}

func (wt *testWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	wt.called = true
	return wt.Reader.(io.WriterTo).WriteTo(w)
}

func TestProxyWriterTo(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	writerTo := &testWriterTo{strings.NewReader(content), false}

	bar := p.AddBar(int64(len(content)))

	var buf bytes.Buffer
	_, err := io.Copy(&buf, bar.ProxyReader(writerTo))
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if !writerTo.called {
		t.Error("WriteTo not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}
