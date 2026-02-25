package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// filename from path
var usedFilenames = make(map[string]int)

func basefilenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "downloaded_file"
	}
	name := path.Base(u.Path)
	if name == "" || name == "/" || name == "." {
		return "downloaded_file"
	}
	return name
}

// ensure unique filename
func uniqueFilename(name string) string {
	count := usedFilenames[name]
	usedFilenames[name]++

	if count == 0 {
		return name
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s_%d%s", base, count, ext)
}

func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func downloadOne(rawURL, outName, outDir string, rateLimitBytesPerSec float64, background bool, logger io.Writer) error {
	// create request with a User-Agent (some sites block the default one)
	client := &http.Client{}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mywget/1.0")
	req.Header.Set("Accept", "*/*")

	fmt.Fprint(logger, "sending request, awaiting response...")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(logger)
		return err
	}
	defer resp.Body.Close()

	fmt.Fprintf(logger, "status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	// check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad request: %s", resp.Status)
	}

	// create output file
	filename := basefilenameFromURL(rawURL)

	if outName != "" {
		filename = outName
	}
	filename = uniqueFilename(filename)

	// ensure output directory
	savePath := filename
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}
		savePath = filepath.Join(outDir, filename)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// download with progress
	progress := NewProgress(resp.ContentLength)

	downloadStart := time.Now()
	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			_, _ = out.Write(buffer[:n])
			progress.Add(n)

			//rate limiting
			if rateLimitBytesPerSec > 0 {
				elapsed := time.Since(downloadStart).Seconds()
				expectedTime := float64(progress.downloaded) / rateLimitBytesPerSec
				if elapsed < expectedTime {
					sleepSeconds := expectedTime - elapsed
					time.Sleep(time.Duration(sleepSeconds * float64(time.Second)))
				}
			}

			if !background && progress.ShouldPrint() {
				progress.Print(os.Stdout)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// final progress print (always)
	if !background {
		progress.Finish(os.Stdout)
	}
	fmt.Fprintf(logger, "Downloaded %d bytes to %s\n", progress.downloaded, savePath)
	fmt.Fprintf(logger, "finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func main() {
	// check args
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <url>")
		return
	}

	// pasre args
	var rawURL string
	var outName string
	var outDir string
	var inputFile string
	var rateLimitBytesPerSec float64
	var background bool
	var mirror bool

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-O=") {
			outName = strings.TrimPrefix(arg, "-O=")
		} else if strings.HasPrefix(arg, "-P=") {
			outDir = strings.TrimPrefix(arg, "-P=")
		} else if strings.HasPrefix(arg, "-i=") {
			inputFile = strings.TrimPrefix(arg, "-i=")
		} else if strings.HasPrefix(arg, "--rate-limit=") {
			val := strings.TrimPrefix(arg, "--rate-limit=")
			if val == "" {
				fmt.Println("Invalid --rate-limit value")
				return
			}
			unit := val[len(val)-1]
			numStr := val
			multiplier := float64(1) //bytes per sec by default

			if unit == 'K' || unit == 'k' {
				multiplier = 1024
				numStr = val[:len(val)-1]
			} else if unit == 'M' || unit == 'm' {
				multiplier = 1024 * 1024
				numStr = val[:len(val)-1]
			}

			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil || num <= 0 {
				fmt.Println("Invalid --rate-limit value:", val)
				return
			}
			rateLimitBytesPerSec = num * multiplier
		} else if arg == "-B" {
			background = true
		} else if arg == "--mirror" {
			mirror = true
		} else if !strings.HasPrefix(arg, "-") {
			//arg is the URL
			rawURL = arg
		}
	}

	if inputFile == "" && rawURL == "" {
		fmt.Println("Usage: go run . [-O=filename] [-P=dir] [--rate-limit=VALUE] [-B] <url>  OR  go run . -i=file")
		return
	}

	var logger io.Writer = os.Stdout
	if background {
		f, err := os.Create("wget-log")
		if err != nil {
			fmt.Println("Cannot create wget-log file:", err)
			return
		}
		defer f.Close()

		logger = f
		fmt.Println(`Output will be written to "wget-log" file.`)
	}

	start := time.Now()
	fmt.Fprintf(logger, "start at %s\n", start.Format("2006-01-02 15:04:05"))

	if mirror {
		if err := MirrorSite(rawURL, logger); err != nil {
			fmt.Fprintln(logger, "Error:", err)
		}
		return
	}

	if inputFile != "" {
		urls, err := readLines(inputFile)
		if err != nil {
			fmt.Fprintln(logger, "Error reading file:", err)
			return
		}

		var wg sync.WaitGroup
		done := make(chan string, len(urls))

		for _, u := range urls {
			wg.Add(1)
			go func(link string) {
				defer wg.Done()
				_ = downloadOne(link, "", outDir, rateLimitBytesPerSec, false, os.Stdout)
				done <- link
			}(u)
		}

		wg.Wait()
		close(done)

		var finished []string
		for link := range done {
			fmt.Println("finished", uniqueFilename(link))
			finished = append(finished, link)
		}
		fmt.Println("Download finished:", finished)
		return
	}

	if err := downloadOne(rawURL, outName, outDir, rateLimitBytesPerSec, background, logger); err != nil {
		fmt.Fprintln(logger, "Error:", err)
	}
}
