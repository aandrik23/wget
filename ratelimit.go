package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Rate Limit enforces a maximum dowload speed
type RateLimiter struct {
	bytesPerSec float64
	start       time.Time // when the download started
}

// NewRateLimiter creates a new RateLimiter with the given bytes per second limit.
func NewRateLimiter(bytesPerSec float64) *RateLimiter {
	return &RateLimiter{
		bytesPerSec: bytesPerSec,
		start:       time.Now(),
	}
}

// Parse strings
func ParseRateLimit(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty rate limit")
	}

	//check last character
	last := s[len(s)-1]
	multiplier := float64(1)
	numberPart := s

	if last == 'k' || last == 'K' {
		multiplier = 1024
		numberPart = s[:len(s)-1]
	} else if last == 'm' || last == 'M' {
		multiplier = 1024 * 1024
		numberPart = s[:len(s)-1]
	}

	num, err := strconv.ParseFloat(numberPart, 64)
	if err != nil || num < 0 {
		return 0, fmt.Errorf("invalid rate limit: %s", s)
	}

	return num * multiplier, nil
}

func (rl *RateLimiter) Throttle(downloadedBytes int) {
	if rl == nil || rl.bytesPerSec <= 0 {
		return // no rate limit
	}

	elapsed := time.Since(rl.start).Seconds()
	if elapsed <= 0 {
		return
	}

	//expected time based on rate limit
	expected := float64(downloadedBytes) / rl.bytesPerSec
	if expected > elapsed {
		sleepSeconds := expected - elapsed
		time.Sleep(time.Duration(sleepSeconds * float64(time.Second)))
	}
}
