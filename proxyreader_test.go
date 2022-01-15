package mpb_test

import (
	"bytes"
	"io"
	"io/ioutil"
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
	p := mpb.New(mpb.WithOutput(ioutil.Discard))

	tReader := &testReader{strings.NewReader(content), false}

	bar := p.AddBar(int64(len(content)), mpb.BarFillerTrim())

	var buf bytes.Buffer
	_, err := io.Copy(&buf, bar.ProxyReader(tReader))
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if !tReader.called {
		t.Error("Read not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}

type testWriterTo struct {
	io.Reader
	wt     io.WriterTo
	called bool
}

func (wt *testWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	wt.called = true
	return wt.wt.WriteTo(w)
}

func TestProxyWriterTo(t *testing.T) {
	p := mpb.New(mpb.WithOutput(ioutil.Discard))

	var reader io.Reader = strings.NewReader(content)
	tReader := &testWriterTo{reader, reader.(io.WriterTo), false}

	bar := p.AddBar(int64(len(content)), mpb.BarFillerTrim())

	var buf bytes.Buffer
	_, err := io.Copy(&buf, bar.ProxyReader(tReader))
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if !tReader.called {
		t.Error("WriteTo not called")
	}

	if got := buf.String(); got != content {
		t.Errorf("Expected content: %s, got: %s\n", content, got)
	}
}
