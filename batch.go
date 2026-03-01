package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

func downloadFromFile(inputFile, outDir string, rateLimit float64, logger io.Writer) error {
	urls, err := readLines(inputFile)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	done := make(chan string, len(urls))

	for _, u := range urls {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			_ = downloadOne(link, "", outDir, rateLimit, false, os.Stdout)
			done <- link
		}(u)
	}

	wg.Wait()
	close(done)

	var finished []string
	for link := range done {
		fmt.Println("finished", filenameFromURL(link))
		finished = append(finished, link)
	}
	fmt.Println("Download finished:", finished)
	return nil
}
