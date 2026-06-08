package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const ddgSearchURL = "https://lite.duckduckgo.com/lite/?q=%s"

// searchAndScrape: searches DuckDuckGo, gets top 3 result URLs, scrapes them.
func searchAndScrape(query string) (string, error) {
	// 1. Perform search
	searchPage, err := fetchHTML(fmt.Sprintf(ddgSearchURL, query))
	if err != nil {
		return "", err
	}
	urls := extractResultLinks(searchPage)
	if len(urls) == 0 {
		return "", fmt.Errorf("no search results found")
	}

	// 2. Scrape top results (max 3, to keep free & fast)
	topUrls := urls
	if len(topUrls) > 3 {
		topUrls = topUrls[:3]
	}
	return scrapeMultiple(topUrls)
}

// extractResultLinks parses DuckDuckGo lite HTML
func extractResultLinks(htmlStr string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		return nil
	}
	var links []string
	doc.Find("a.result-link").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists && strings.HasPrefix(href, "http") {
			links = append(links, href)
		}
	})
	// fallback: look for any external link
	if len(links) == 0 {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists &&
				(strings.HasPrefix(href, "http") || strings.HasPrefix(href, "https")) &&
				!strings.Contains(href, "duckduckgo.com") {
				links = append(links, href)
			}
		})
	}
	return links
}
