package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// MirrorSire downloads only the first page into a folder
func MirrorSite(rawURL string, logger io.Writer) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	//find directory name
	domainDir := u.Host
	if domainDir == "" {
		return fmt.Errorf("url has no host")
	}

	// create a domain folder
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return fmt.Errorf("cannot create domain folder: %w", err)
	}

	// decide local path for url
	savePath := mirrorSavePath(domainDir, u)

	// stdout or wget-log
	fmt.Fprintf(logger, "saving file to: %s\n", savePath)
	return downloadFileSimple(u.String(), savePath, logger)
}

// maps a URL to a local filepath under domainDir
// "/" or paths ending with "/" become ".../index.html"
func mirrorSavePath(domainDir string, u *url.URL) string {
	p := u.Path
	if p == "" || strings.HasSuffix(p, "/") {
		p = p + "index.html"
	}

	//make sure path is inside domainDir
	p = strings.TrimPrefix(p, "/")
	return filepath.Join(domainDir, filepath.FromSlash(p))
}

// download urlStr into savepath
func downloadFileSimple(urlStr, savePath string, logger io.Writer) error {
	// cretate folders for nested files
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		return fmt.Errorf("cannot cretae folders: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "wget/1.0")
	req.Header.Set("Accept", "*/*")

	fmt.Fprint(logger, "sending request, awaiting response...")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Fprintf(logger, "status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	return nil
}
