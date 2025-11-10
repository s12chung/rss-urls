package item

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Author      string `xml:"author"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type HTMLMetadata struct {
	Title       string
	Description string
}

func FromURL(u string) (Item, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return Item{}, fmt.Errorf("url parse error: %w", err)
	}

	resp, err := http.Get(u)
	if err != nil {
		return Item{}, fmt.Errorf("fetch error: %w", err)
	}
	defer func() {
		err2 := resp.Body.Close()
		if err == nil {
			err = err2
		}
	}()

	meta, err := contentMeta(resp, parsedURL)
	if err != nil {
		return Item{}, err
	}
	return Item{
		Title:       meta.Title,
		Link:        u,
		Description: meta.Description,
		Author:      strings.TrimPrefix(parsedURL.Host, "www."),
		PubDate:     time.Now().Format(time.RFC1123Z),
		GUID:        u,
	}, err
}

func contentMeta(resp *http.Response, parsedURL *url.URL) (*HTMLMetadata, error) {
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))

	if strings.Contains(contentType, "text/html") {
		return traverseHTML(resp.Body)
	} else if strings.Contains(contentType, "application/pdf") {
		return pdfMeta(parsedURL), nil
	}

	return nil, fmt.Errorf("unsupported content type: %s", contentType)
}

func traverseHTML(body io.ReadCloser) (*HTMLMetadata, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("html parse error: %w", err)
	}

	meta := &HTMLMetadata{}
	traverse(doc, meta)

	if meta.Title == "" {
		meta.Title = "No title"
	}
	if meta.Description == "" {
		meta.Description = "No description"
	}
	return meta, nil
}

func pdfMeta(parsedURL *url.URL) *HTMLMetadata {
	filename := parsedURL.Path
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}
	if filename == "" {
		filename = "document.pdf"
	}

	return &HTMLMetadata{
		Title:       filename,
		Description: "A PDF File",
	}
}

func traverse(n *html.Node, meta *HTMLMetadata) {
	if n.Type == html.ElementNode {
		if n.Data == "title" && meta.Title == "" {
			if n.FirstChild != nil {
				meta.Title = n.FirstChild.Data
			}
		}
		if n.Data == "meta" && meta.Description == "" {
			var name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name", "property":
					if attr.Val == "description" || attr.Val == "og:description" {
						name = attr.Val
					}
				case "content":
					content = attr.Val
				}
			}
			if name != "" && content != "" {
				meta.Description = content
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if meta.Title != "" && meta.Description != "" {
			return
		}
		traverse(c, meta)
	}
}
