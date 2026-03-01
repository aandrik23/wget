package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
)

// Recursive mirror using BFS
func MirrorSiteStep2(rawURL string, logger io.Writer) error {
	startURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if startURL.Scheme == "" {
		startURL.Scheme = "https"
	}

	domain := startURL.Host
	if domain == "" {
		return fmt.Errorf("url has no host")
	}

	// create root domain folder
	if err := os.MkdirAll(domain, 0755); err != nil {
		return err
	}

	visited := make(map[string]bool)
	queue := []string{startURL.String()}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		u, err := url.Parse(current)
		if err != nil {
			continue
		}

		// Only same host
		if u.Host != domain {
			continue
		}

		savePath := mirrorSavePath(domain, u)
		fmt.Fprintf(logger, "saving file to: %s\n", savePath)

		// Download file
		if err := downloadFileSimple(current, savePath, logger); err != nil {
			continue
		}

		// Only parse HTML files
		if !isHTML(savePath) {
			continue
		}

		// Read saved HTML
		b, err := os.ReadFile(savePath)
		if err != nil {
			continue
		}

		links, err := ExtractLinksFromBytes(b, u)
		if err != nil {
			continue
		}

		for _, l := range links {
			linkURL, err := url.Parse(l.URL)
			if err != nil {
				continue
			}

			if linkURL.Host == domain {
				queue = append(queue, linkURL.String())
			}
		}
	}

	return nil
}

// Check if file is HTML
func isHTML(path string) bool {
	if len(path) < 5 {
		return true
	}
	if path[len(path)-5:] == ".html" {
		return true
	}
	if path[len(path)-4:] == ".htm" {
		return true
	}
	return false
}
