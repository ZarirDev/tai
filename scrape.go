package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func fetchHTML(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; AIBot/1.0)")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	// Remove script, style, and other non-content tags
	doc.Find("script, style, nav, footer, header, aside").Remove()
	// Extract text
	text := doc.Find("body").Text()
	// Clean up whitespace
	text = strings.Join(strings.Fields(text), " ")
	return text, nil
}

func scrapeMultiple(urls []string) (string, error) {
	var allText strings.Builder
	for _, u := range urls {
		fmt.Fprintf(os.Stderr, "Scraping: %s\n", u)
		text, err := fetchHTML(u)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to scrape %s: %v\n", u, err)
			continue
		}
		// Truncate very long pages (e.g., > 10k chars) to avoid huge prompts
		if len(text) > 10000 {
			text = text[:10000] + "... (truncated)"
		}
		allText.WriteString(fmt.Sprintf("\n--- Content from %s ---\n%s\n", u, text))
	}
	if allText.Len() == 0 {
		return "", fmt.Errorf("no content could be scraped")
	}
	return allText.String(), nil
}
