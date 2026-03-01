package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// check args
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . [flags] <url>")
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
		fmt.Println("Usage: go run . [-O=filename] [-P=dir] [-i=file] [--rate-limit=VALUE] [-B] [--mirror] <url>")
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

	if inputFile != "" {
		if err := downloadFromFile(inputFile, outDir, rateLimitBytesPerSec, logger); err != nil {
			fmt.Fprintln(logger, "Error:", err)
		}
		return
	}

	if mirror {
		if err := MirrorSiteStep2(rawURL, logger); err != nil {
			fmt.Fprintln(logger, "Error:", err)
		}
		return
	}

	if err := downloadOne(rawURL, outName, outDir, rateLimitBytesPerSec, background, logger); err != nil {
		fmt.Fprintln(logger, "Error:", err)
	}
}
