package mpb_test

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/vbauerster/mpb"
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

	reader := &testReader{Reader: strings.NewReader(content)}

	total := len(content)
	bar := p.AddBar(100, mpb.TrimSpace())

	written, err := io.Copy(ioutil.Discard, bar.ProxyReader(reader))
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if !reader.called {
		t.Error("Read not called")
	}

	if written != int64(total) {
		t.Errorf("Expected written: %d, got: %d\n", total, written)
	}
}
