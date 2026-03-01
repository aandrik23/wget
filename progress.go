package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Progress struct {
	totalBytes   int64 // total bytes to download (-1 if unknown)
	downloaded   int64
	start        time.Time
	lastPrint    time.Time
	printEvery   time.Duration
	barWidth     int
	unknownTotal bool
}

// NewProgress creates a new progress tracker.
func NewProgress(totalBytes int64) *Progress {
	p := &Progress{
		totalBytes: totalBytes,
		start:      time.Now(),
		lastPrint:  time.Now(),
		printEvery: 100 * time.Millisecond,
		barWidth:   30,
	}
	if totalBytes <= 0 {
		p.unknownTotal = true
	}
	return p
}

// Add increases downloaded bytes.
func (p *Progress) Add(n int) {
	p.downloaded += int64(n)
}

// ShouldPrint returns true if enough time passed since last print.
func (p *Progress) ShouldPrint() bool {
	return time.Since(p.lastPrint) >= p.printEvery
}

// prints one progress line (updates same terminal line).
func (p *Progress) Print(w io.Writer) {
	p.lastPrint = time.Now()

	elapsed := time.Since(p.start).Seconds()
	if elapsed < 0.001 {
		elapsed = 0.001
	}

	// average speed since start (bytes/sec)
	speedBytesPerSec := float64(p.downloaded) / elapsed
	speedMiBPerSec := speedBytesPerSec / (1024 * 1024)

	downloadedMiB := bytesToMiB(p.downloaded)

	// If total size is unknown, we can only show downloaded + speed.
	if p.unknownTotal {
		fmt.Fprintf(w, "\rDownloaded %.2f MiB  %.2f MiB/s\033[K", downloadedMiB, speedMiBPerSec)
		return
	}

	// Total is known: show percentage, bar, ETA.
	totalMiB := bytesToMiB(p.totalBytes)
	percent := float64(p.downloaded) / float64(p.totalBytes) * 100
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	bar := buildBar(percent, p.barWidth)

	remainingBytes := float64(p.totalBytes - p.downloaded)
	etaSeconds := 0.0
	if speedBytesPerSec > 0 && remainingBytes > 0 {
		etaSeconds = remainingBytes / speedBytesPerSec
	}

	fmt.Fprintf(
		w,
		"\r%.2f MiB / %.2f MiB %s %.2f%%  %.2f MiB/s  %ds\033[K",
		downloadedMiB, totalMiB, bar, percent, speedMiBPerSec, int(etaSeconds),
	)
}

// Finish prints the final progress line and moves to a new line.
func (p *Progress) Finish(w io.Writer) {
	fmt.Fprintln(w)
}

// bytesToMiB converts bytes to MiB.
func bytesToMiB(b int64) float64 {
	return float64(b) / (1024 * 1024)
}

// buildBar creates the progress bar string.
func buildBar(percent float64, barWidth int) string {
	completedBars := int((percent / 100) * float64(barWidth))
	if completedBars < 0 {
		completedBars = 0
	}
	if completedBars > barWidth {
		completedBars = barWidth
	}
	remaining := barWidth - completedBars
	return "[" + strings.Repeat("=", completedBars) + ">" + strings.Repeat("-", remaining) + "]"
}
