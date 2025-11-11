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
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type HTMLMetadata struct {
	Title       string
	Description string
	FinalURL    string
}

type Anchor struct {
	Href string
	Text string
}

func (a Anchor) HTML() string { return fmt.Sprintf(`<a href="%s">%s</a>`, a.Href, a.Text) }

func FromURL(u string) (Item, error) {
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

	meta, err := contentMeta(resp)
	if err != nil {
		return Item{}, err
	}

	simpleHost := strings.TrimPrefix(resp.Request.URL.Host, "www.")
	return Item{
		Title:       meta.Title,
		Link:        meta.FinalURL,
		Description: Anchor{Href: u, Text: simpleHost}.HTML() + " - " + meta.Description,
		PubDate:     time.Now().Format(time.RFC1123Z),
		GUID:        meta.FinalURL,
	}, err
}

func contentMeta(resp *http.Response) (*HTMLMetadata, error) {
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	cleanHost := removeSubdomain(resp.Request.URL.Host)

	var meta *HTMLMetadata
	var err error
	if strings.Contains(contentType, "application/pdf") {
		meta = pdfMeta(resp.Request.URL)
	} else if strings.Contains(contentType, "text/html") {
		meta, err = traverseHTML(resp.Body)
		if err != nil {
			return nil, err
		}
		meta.Title = htmlTitle(meta.Title, cleanHost)
	} else {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	meta.FinalURL = cleanURL(resp.Request.URL, cleanHost)
	return meta, nil
}

var urlParams = map[string][]string{
	"youtube.com": {"v", "list"},
}

func cleanURL(u *url.URL, cleanHost string) string {
	params, found := urlParams[cleanHost]
	if found {
		u.RawQuery = limitParams(u, params).Encode()
	}
	return u.String()
}

func limitParams(u *url.URL, allowed []string) url.Values {
	filtered := url.Values{}
	q := u.Query()
	for _, key := range allowed {
		if val := q.Get(key); val != "" {
			filtered.Set(key, val)
		}
	}
	return filtered
}

func pdfMeta(u *url.URL) *HTMLMetadata {
	filename := u.Path
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}

	return &HTMLMetadata{
		Title:       "📑 " + filename,
		Description: "A PDF File",
	}
}

var htmlEmojis = map[string]string{
	"youtube.com":  "📺",
	"substack.com": "🟧",
}

func htmlTitle(title, cleanHost string) string {
	prefix := htmlEmojis[cleanHost]
	if prefix != "" {
		prefix = prefix + " "
	}
	return prefix + title
}

func traverseHTML(body io.ReadCloser) (*HTMLMetadata, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	meta := &HTMLMetadata{}
	traverseNode(doc, meta)

	if meta.Title == "" {
		meta.Title = "No title"
	}
	if meta.Description == "" {
		meta.Description = "No description"
	}
	return meta, nil
}

func traverseNode(n *html.Node, meta *HTMLMetadata) {
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
		traverseNode(c, meta)
	}
}
