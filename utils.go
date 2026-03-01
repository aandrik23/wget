package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func readLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func filenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "downloaded_file"
	}

	name := path.Base(u.Path)
	if name == "" || name == "/" || name == "." {
		name = "downloaded_file"
	}

	// If name is empty/generic, try to add query info for uniqueness
	if q := u.RawQuery; q != "" {
		// safe-ish suffix
		suffix := strings.NewReplacer("&", "_", "=", "_", "?", "_").Replace(q)
		if len(suffix) > 40 {
			suffix = suffix[:40]
		}
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		return fmt.Sprintf("%s_%s%s", base, suffix, ext)
	}

	return name
}
