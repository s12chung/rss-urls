package item

import (
	"fmt"
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
		return Item{}, fmt.Errorf("parse error: %w", err)
	}
	meta, err := traverseHTML(u)
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

func traverseHTML(u string) (*HTMLMetadata, error) {
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %w", err)
	}
	defer func() {
		err2 := resp.Body.Close()
		if err == nil {
			err = err2
		}
	}()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	meta := &HTMLMetadata{}
	traverse(doc, meta)

	if meta.Title == "" {
		meta.Title = u
	}
	if meta.Description == "" {
		meta.Description = "No description"
	}
	return meta, nil
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
