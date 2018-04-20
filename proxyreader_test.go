package mpb_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

	written, err := io.Copy(ioutil.Discard, preader)
	if err != nil {
		t.Errorf("Error copying from reader: %+v\n", err)
	}

	p.Wait()

	if written != int64(total) {
		t.Errorf("Expected written: %d, got: %d\n", total, written)
	}

	// underlying reader is not Closer
	err = preader.Close()
	if err != nil {
		t.Errorf("Expected nil error, got: %+v\n", err)
	}
}

func TestProxyReaderCloser(t *testing.T) {
	var buf bytes.Buffer
	p := mpb.New(mpb.WithOutput(&buf))

	ts := setupTestHttpServer(content)
	defer ts.Close()

	url := ts.URL + "/test"
	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("Test server get failure: %s\n", url)
	}

	total := resp.ContentLength
	bar := p.AddBar(total, mpb.BarTrim())
	reader := bar.ProxyReader(resp.Body)

	// calling reader.Close() will call resp.Body.Close() implicitly
	err = reader.Close()
	if err != nil {
		t.Logf("Error closing resp.Body over reader.Close: %+v\n", err)
		t.FailNow()
	}

	// reading from closed resp.Body
	_, err = io.Copy(ioutil.Discard, reader)
	if err == nil {
		t.Error("Expected read on closed response body error!")
	}
}

func setupTestHttpServer(content string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, content)
	})
	return httptest.NewServer(mux)
}
