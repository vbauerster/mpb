package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/vbauerster/mpb"
)

func main() {
	log.SetOutput(os.Stderr)

	url1 := "https://homebrew.bintray.com/bottles/youtube-dl-2016.12.12.sierra.bottle.tar.gz"
	url2 := "https://homebrew.bintray.com/bottles/libtiff-4.0.7.sierra.bottle.tar.gz"

	var wg sync.WaitGroup
	p := mpb.New().SetWidth(60)

	for i, url := range [...]string{url1, url2} {
		wg.Add(1)
		name := fmt.Sprintf("url%d:", i+1)
		go download(&wg, p, name, url)
	}

	wg.Wait()
	p.Stop()
	fmt.Println("Finished")
}

func download(wg *sync.WaitGroup, p *mpb.Progress, name, url string) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("%s: %v", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status: %s", resp.Status)
		log.Printf("%s: %v", name, err)
		return
	}

	size := resp.ContentLength

	// create dest
	destName := filepath.Base(url)
	dest, err := os.Create(destName)
	if err != nil {
		err = fmt.Errorf("Can't create %s: %v", destName, err)
		log.Printf("%s: %v", name, err)
		return
	}

	// create bar with appropriate decorators
	bar := p.AddBar(size).
		PrependName(name, 0, 0).
		PrependCounters("%3s / %3s", mpb.UnitBytes, 18, mpb.DwidthSync|mpb.DextraSpace).
		AppendETA(5, mpb.DwidthSync)

	// create proxy reader
	reader := bar.ProxyReader(resp.Body)
	// and copy from reader
	_, err = io.Copy(dest, reader)

	if closeErr := dest.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		log.Printf("%s: %v", name, err)
	}
}
