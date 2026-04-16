package mpb_test

import (
	"bytes"
	"errors"
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
	r      io.Reader
	called bool
}

func (r *testReader) Read(p []byte) (n int, err error) {
	r.called = true
	return r.r.Read(p)
}

func TestProxyReader(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	tr := &testReader{strings.NewReader(content), false}
	_, err := io.Copy(&buf, bar.ProxyReader(tr))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	if !tr.called {
		t.Error("Read not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
	p.Wait()
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
	bar := p.New(int64(len(content)), mpb.NopStyle())

	tr := &testReadCloser{strings.NewReader(content), false}
	rc := bar.ProxyReader(tr)
	_, err := io.Copy(io.Discard, rc)
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}
	_ = rc.Close()
	if !tr.called {
		t.Error("Close not called")
	}
	p.Wait()
}

type testReaderWriterTo struct {
	r      io.Reader
	called bool
}

func (r *testReaderWriterTo) Read(p []byte) (n int, err error) {
	return 0, errors.New("unexpected")
}

func (r *testReaderWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	r.called = true
	return r.r.(io.WriterTo).WriteTo(w)
}

func TestProxyReaderWriterTo(t *testing.T) {
	p := mpb.New(mpb.WithOutput(io.Discard))
	bar := p.New(int64(len(content)), mpb.NopStyle())

	var buf bytes.Buffer
	tr := &testReaderWriterTo{strings.NewReader(content), false}
	_, err := io.Copy(&buf, bar.ProxyReader(tr))
	if err != nil {
		t.Errorf("io.Copy: %s\n", err.Error())
	}

	if !tr.called {
		t.Error("WriteTo not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
	p.Wait()
}
