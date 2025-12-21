package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	// check args
	var rawURL string
	var outputName string
	var outputDir string

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-O=") {
			outputName = strings.TrimPrefix(arg, "-O=")
		} else if strings.HasPrefix(arg, "-P=") {
			outputDir = strings.TrimPrefix(arg, "-P=")
		} else if !strings.HasPrefix(arg, "-") {
			rawURL = arg
		}
	}

	if rawURL == "" {
		fmt.Println("Usage: go run . [-O=filename] [-P=dir] URL")
		return
	}

	//start time
	start := time.Now()
	fmt.Printf("start at %s\n", start.Format("2006-01-02 15:04:05"))

	// url parse
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Invalid URL:", err)
		return
	}

	filename := path.Base(parsedURL.Path)
	if outputName != "" {
		filename = outputName
	}
	if filename == "" || filename == "." || filename == "/" {
		filename = "downloaded_file"
	}

	savePath := filename
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
		savePath = path.Join(outputDir, filename)
	}

	// HTTP request
	fmt.Print("sending request, awaiting response...")

	client := &http.Client{}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("User-Agent", "mywget/1.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while making GET request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status:", resp.Status)
		return
	}

	// size info
	size := resp.ContentLength
	unknownSize := size <= 0

	mb := float64(size) / (1024 * 1024)
	fmt.Printf("content size: %d bytes [~%.2fMB]\n", size, mb)
	fmt.Printf("saving file to: ./%s\n\n", savePath)

	// create file
	outFile, err := os.Create(savePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	// download loop
	buffer := make([]byte, 32*1024)
	var downloaded int64

	downloadStart := time.Now()
	lastPrint := time.Now()

	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			outFile.Write(buffer[:n])
			downloaded += int64(n)

			if time.Since(lastPrint) >= 100*time.Millisecond {
				lastPrint = time.Now()

				elapsed := time.Since(downloadStart).Seconds()
				if elapsed < 0.001 {
					elapsed = 0.001
				}

				speedBytes := float64(downloaded) / elapsed
				speedMiB := speedBytes / (1024 * 1024)

				if unknownSize {
					fmt.Printf("\r\033[KDownloaded %.2f MiB  %.2f MiB/s",
						float64(downloaded)/(1024*1024), speedMiB)
				} else {
					percent := float64(downloaded) / float64(size) * 100
					if percent > 100 {
						percent = 100
					}

					barWidth := 50
					filled := int(percent / 100 * float64(barWidth))
					if filled > barWidth {
						filled = barWidth
					}

					bar := "[" + strings.Repeat("=", filled) +
						">" + strings.Repeat("-", barWidth-filled) + "]"

					remaining := float64(size-downloaded) / speedBytes
					if remaining < 0 {
						remaining = 0
					}

					fmt.Printf(
						"\r\033[K%.2f MiB / %.2f MiB %s %.2f%%  %.2f MiB/s  %ds",
						float64(downloaded)/(1024*1024),
						float64(size)/(1024*1024),
						bar,
						percent,
						speedMiB,
						int(remaining),
					)
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error while downloading:", err)
			return
		}
	}

	fmt.Println()
	fmt.Printf("Downloaded %d bytes to %s\n", downloaded, savePath)
	fmt.Printf("finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
