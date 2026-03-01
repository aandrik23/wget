package main

import (
	"bytes"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ExtractLinksFromBytes(pageURL *url.URL, body []byte) []string {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	return ExtractLinksFromHTML(pageURL, doc)
}

// ExtractLinksFromHTML walks the HTML tree and returns absolute URLs found in:
// - a[href]
// - link[href]
// - img[src]
// - script[src]        <-- IMPORTANT for many sites
// - source[src]        <-- useful for audio/video
// - audio[src], video[src] (optional but harmless)
func ExtractLinksFromHTML(pageURL *url.URL, n *html.Node) []string {
	var out []string

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)

			// Decide which attribute we want to read based on the tag
			var attrName string
			switch tag {
			case "a", "link":
				attrName = "href"
			case "img", "script", "source", "audio", "video":
				attrName = "src"
			}

			if attrName != "" {
				for _, a := range node.Attr {
					if strings.ToLower(a.Key) != attrName {
						continue
					}
					raw := strings.TrimSpace(a.Val)
					if raw == "" {
						continue
					}

					// Ignore anchors and non-http(s) stuff
					if strings.HasPrefix(raw, "#") ||
						strings.HasPrefix(raw, "mailto:") ||
						strings.HasPrefix(raw, "javascript:") ||
						strings.HasPrefix(raw, "data:") {
						continue
					}

					ref, err := url.Parse(raw)
					if err != nil {
						continue
					}

					abs := pageURL.ResolveReference(ref)
					out = append(out, abs.String())
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return out
}
