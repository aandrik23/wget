package main

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type FoundLink struct {
	URL      string // absolute URL
	Tag      string // a / img / link
	Attr     string // href / src
	Original string // original attribute value (as found)
}

// ExtractLinksFromHTML parses HTML and extracts URLs from:
// - a[href]
// - img[src]
// - link[href]
func ExtractLinksFromHTML(r io.Reader, base *url.URL) ([]FoundLink, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("html parse: %w", err)
	}

	var out []FoundLink

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				if v, ok := getAttr(n, "href"); ok {
					addLink(&out, base, "a", "href", v)
				}
			case "img":
				if v, ok := getAttr(n, "src"); ok {
					addLink(&out, base, "img", "src", v)
				}
			case "link":
				if v, ok := getAttr(n, "href"); ok {
					addLink(&out, base, "link", "href", v)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return out, nil
}

// Helper: parse bytes as reader
func ExtractLinksFromBytes(b []byte, base *url.URL) ([]FoundLink, error) {
	return ExtractLinksFromHTML(bytes.NewReader(b), base)
}

func getAttr(n *html.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			val := strings.TrimSpace(a.Val)
			if val == "" {
				return "", false
			}
			return val, true
		}
	}
	return "", false
}

func addLink(out *[]FoundLink, base *url.URL, tag, attr, raw string) {
	// Ignore anchors and javascript
	if strings.HasPrefix(raw, "#") {
		return
	}
	if strings.HasPrefix(strings.ToLower(raw), "javascript:") {
		return
	}
	if strings.HasPrefix(strings.ToLower(raw), "mailto:") {
		return
	}
	if strings.HasPrefix(strings.ToLower(raw), "tel:") {
		return
	}
	if strings.HasPrefix(strings.ToLower(raw), "data:") {
		return
	}

	u, err := url.Parse(raw)
	if err != nil {
		return
	}

	abs := base.ResolveReference(u)

	*out = append(*out, FoundLink{
		URL:      abs.String(),
		Tag:      tag,
		Attr:     attr,
		Original: raw,
	})
}
