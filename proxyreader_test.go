package mpb_test

import (
	"bytes"
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

func TestProxyReader(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(mpb.WithOutput(&buf))

	reader := strings.NewReader(content)

	total := len(content)
	bar := p.AddBar(100, mpb.BarTrim())
	preader := bar.ProxyReader(reader)

	if _, ok := preader.(io.Closer); !ok {
		t.Error("type assertion to io.Closer is not ok")
	}

	written, err := io.Copy(ioutil.Discard, preader)
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if written != int64(total) {
		t.Errorf("Expected written: %d, got: %d\n", total, written)
	}
}
