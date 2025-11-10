package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/s12chung/rss-urls/pkg/item"
)

const rssFile = "rss.xml"

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string      `xml:"title"`
	Link        string      `xml:"link"`
	Description string      `xml:"description"`
	Items       []item.Item `xml:"item"`
}

var (
	defaultRSS = RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       getEnvOrDefault("RSS_TITLE", "My Feed"),
			Link:        getEnvOrDefault("RSS_LINK", "http://localhost"),
			Description: getEnvOrDefault("RSS_DESCRIPTION", "Generated feed"),
		},
	}
	ErrItemExists = fmt.Errorf("item already exists")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: program <url>")
		os.Exit(1)
	}

	url := os.Args[1]

	if err := run(url); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(url string) error {
	// Extract metadata
	i, err := item.FromURL(url)
	if err != nil {
		return err
	}

	// Add to feed
	rss, err := appendRSSItem(i)
	if err != nil {
		if err == ErrItemExists {
			return nil
		}
		return err
	}

	// Write feed
	return writeRSS(rss)
}

func appendRSSItem(i item.Item) (RSS, error) {
	// Load or create feed
	var rss RSS
	data, err := os.ReadFile(rssFile)
	if err == nil {
		if err := xml.Unmarshal(data, &rss); err != nil {
			return RSS{}, err
		}
	} else {
		rss = defaultRSS
	}

	// Check if URL already exists
	for _, existing := range rss.Channel.Items {
		if existing.Link == i.Link {
			return rss, ErrItemExists
		}
	}
	rss.Channel.Items = append([]item.Item{i}, rss.Channel.Items...)

	b, err := json.MarshalIndent(&i, "", "  ")
	if err != nil {
		return RSS{}, err
	}
	fmt.Println("Adding new item: " + string(b))
	return rss, nil
}

func writeRSS(rss RSS) error {
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	err = os.WriteFile(rssFile, []byte(xml.Header+string(output)), 0644)
	if err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	fmt.Println("Added to " + rssFile)
	return nil
}
