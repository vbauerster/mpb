package mpb_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/vbauerster/mpb/v8"
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

	tr := &testReader{strings.NewReader(content), false}

	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	_, err := io.Copy(&buf, bar.ProxyReader(tr))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	p.Wait()

	if !tr.called {
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

	tr := &testReadCloser{strings.NewReader(content), false}

	bar := p.New(int64(len(content)), mpb.NopStyle())

	rc := bar.ProxyReader(tr)
	_, err := io.Copy(io.Discard, rc)
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}
	_ = rc.Close()

	p.Wait()

	if !tr.called {
		t.Error("Close not called")
	}
}

type testReaderWriterTo struct {
	io.Reader
	called bool
}

func (r *testReaderWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	r.called = true
	return r.Reader.(io.WriterTo).WriteTo(w)
}

func TestProxyReaderWriterTo(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))

	tr := &testReaderWriterTo{strings.NewReader(content), false}

	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	_, err := io.Copy(&buf, bar.ProxyReader(tr))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	p.Wait()

	if !tr.called {
		t.Error("WriteTo not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}
