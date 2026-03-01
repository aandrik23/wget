package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func downloadOne(rawURL, outName, outDir string, rateLimitBytesPerSec float64, background bool, logger io.Writer) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "wget/1.0")
	req.Header.Set("Accept", "*/*")

	fmt.Fprint(logger, "sending request, awaiting response...")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Fprintf(logger, "status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	filename := filenameFromURL(rawURL)
	if outName != "" {
		filename = outName
	}

	savePath := filename
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}
		savePath = filepath.Join(outDir, filename)
	}

	fmt.Fprintf(logger, "saving file to: ./%s\n", savePath)

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	progress := NewProgress(resp.ContentLength)

	downloadStart := time.Now()
	buffer := make([]byte, 32*1024)

	for {
		n, rerr := resp.Body.Read(buffer)
		if n > 0 {
			_, _ = out.Write(buffer[:n])
			progress.Add(n)

			// rate limiting
			if rateLimitBytesPerSec > 0 {
				elapsed := time.Since(downloadStart).Seconds()
				expected := float64(progress.downloaded) / rateLimitBytesPerSec
				if elapsed < expected {
					sleepSeconds := expected - elapsed
					time.Sleep(time.Duration(sleepSeconds * float64(time.Second)))
				}
			}

			// progress output (no bar in background mode)
			if !background && progress.ShouldPrint() {
				progress.Print(os.Stdout)
			}
		}

		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}

	if !background {
		progress.Print(os.Stdout)
		progress.Finish(os.Stdout)
	}

	fmt.Fprintf(logger, "Downloaded %d bytes to %s\n", progress.downloaded, savePath)
	fmt.Fprintf(logger, "finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}
