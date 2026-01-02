package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func main() {
	// check args
	var rawURL string
	var outputName string
	var outputDir string
	var rateLimitBytes float64
	var background bool

	for _, arg := range os.Args[1:] {
		if arg == "-B" {
			background = true //background mode flag
		} else if strings.HasPrefix(arg, "-O=") {
			outputName = strings.TrimPrefix(arg, "-O=")
		} else if strings.HasPrefix(arg, "-P=") {
			outputDir = strings.TrimPrefix(arg, "-P=")
		} else if strings.HasPrefix(arg, "--rate-limit=") {
			val := strings.TrimPrefix(arg, "--rate-limit=")
			if val == "" {
				fmt.Println("Invalid rate limit value")
				return
			}
			unit := val[len(val)-1] //last character
			numStr := val

			multiplier := float64(1) //default = bytes/sec

			if unit == 'k' || unit == 'K' {
				multiplier = 1024
				numStr = val[:len(val)-1]
			} else if unit == 'm' || unit == 'M' {
				multiplier = 1024 * 1024
				numStr = val[:len(val)-1]
			}

			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil || num <= 0 {
				fmt.Println("Invalid --rate-limit number:", val)
				return
			}

			rateLimitBytes = num * multiplier

		} else if !strings.HasPrefix(arg, "-") {
			rawURL = arg
		}
	}

	if rawURL == "" {
		fmt.Println("Usage: go run . [-B] [--rate-limit=VALUE] [-P=dir] [-O=filename] URL")
		return
	}

	//setup logger for background mode
	var logger io.Writer = os.Stdout
	if background {
		f, err := os.Create("wget-log")
		if err != nil {
			fmt.Println("Cannot create wget-log file:", err)
			return
		}
		defer f.Close()

		logger = f
		fmt.Println(`Output will be witten to wget-log file.`)
	}
	// time
	start := time.Now()
	fmt.Fprintf(logger, "start at %s\n", start.Format("2006-01-02 15:04:05"))

	// check if string is a valid URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Invalid URL:", err)
		return
	}

	// find filename from URL path
	filename := path.Base(parsedURL.Path)
	if outputName != "" {
		filename = outputName
	}
	if filename == "" || filename == "/" || filename == "." {
		filename = "downloaded_file"
	}

	// send http GET request
	fmt.Fprint(logger, "sending request, awaiting response...")

	// http request with custom User-Agent
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

	fmt.Fprintf(logger, "status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	// check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(logger, "Request failed with status:", resp.Status)
		return
	}

	// length
	size := resp.ContentLength
	unknownSize := size <= 0

	mb := float64(size) / (1024 * 1024)
	fmt.Fprintf(logger, "content size: %d bytes [~%.2fMB]\n", size, mb)

	// save path
	savePath := filename
	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Fprintln(logger, "Error creating directory:", err)
			return
		}
		savePath = path.Join(outputDir, filename)
	}

	fmt.Fprintf(logger, "saving file to: ./%s\n", savePath)

	outFile, err := os.Create(savePath)
	if err != nil {
		fmt.Fprintln(logger, "Error creating file:", err)
		return
	}
	defer outFile.Close()

	// write response body to file
	buffer := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64 = 0

	downloadStart := time.Now()
	lastPrint := time.Now()

	for {
		n, err := resp.Body.Read(buffer)

		if n > 0 {
			outFile.Write(buffer[:n])
			downloaded += int64(n)

			//rate time logic
			if rateLimitBytes > 0 {
				elapsedAll := time.Since(downloadStart).Seconds()
				if elapsedAll > 0 {
					expectedTime := float64(downloaded) / rateLimitBytes
					if expectedTime > elapsedAll {
						sleepSeconds := expectedTime - elapsedAll
						time.Sleep(time.Duration(sleepSeconds * float64(time.Second)))
					}
				}
			}

			// progress bar
			if !background && time.Since(lastPrint) >= 100*time.Millisecond {
				lastPrint = time.Now()

				elapsed := time.Since(downloadStart).Seconds()
				if elapsed < 0.001 {
					elapsed = 0.001
				}

				speedBytesPerSec := float64(downloaded) / elapsed
				speedMiBPerSec := speedBytesPerSec / (1024 * 1024)

				if unknownSize {
					downloadedMB := float64(downloaded) / (1024 * 1024)
					fmt.Printf("\rDownloaded %.2f MiB  %.2f MiB/s", downloadedMB, speedMiBPerSec)
				} else {
					percent := float64(downloaded) / float64(size) * 100
					if percent < 0 {
						percent = 0
					}
					if percent > 100 {
						percent = 100
					}

					downloadedMB := float64(downloaded) / (1024 * 1024)
					totalMB := float64(size) / (1024 * 1024)

					barWidth := 50
					completedBars := int((percent / 100) * float64(barWidth))
					if completedBars < 0 {
						completedBars = 0
					}
					if completedBars > barWidth {
						completedBars = barWidth
					}

					bar := "[" + strings.Repeat("=", completedBars) +
						">" + strings.Repeat("-", barWidth-completedBars) + "]"

					remainingBytes := float64(size - downloaded)
					etaSeconds := 0.0
					if speedBytesPerSec > 0 && remainingBytes > 0 {
						etaSeconds = remainingBytes / speedBytesPerSec
					}

					fmt.Printf("\r%.2f MiB / %.2f MiB %s %.2f%%  %.2f MiB/s  %ds",
						downloadedMB, totalMB, bar, percent, speedMiBPerSec, int(etaSeconds))
				}
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintln(logger, "Error while downloading:", err)
			return
		}
	}

	fmt.Fprintln(logger) // new line after progress bar

	fmt.Fprintf(logger, "Downloaded %d bytes to %s\n", downloaded, savePath)
	fmt.Fprintf(logger, "finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
