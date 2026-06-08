package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// searchAndScrape uses DDG HTML version, extracts top 3 result URLs, and scrapes them.
func searchAndScrape(query string) (string, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	html, err := fetchHTML(searchURL)
	if err != nil {
		return "", fmt.Errorf("search fetch: %w", err)
	}

	urls := extractResultURLs(html)
	if len(urls) == 0 {
		return "", fmt.Errorf("no results found")
	}
	if len(urls) > 3 {
		urls = urls[:3]
	}
	return scrapeMultiple(urls)
}

// extractResultURLs parses DDG HTML result links.
func extractResultURLs(htmlStr string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		return nil
	}
	var urls []string
	doc.Find("a.result__a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		// DDG uses relative redirect URLs like "/url?q=ACTUAL_URL"
		if strings.HasPrefix(href, "/url?q=") {
			if u, err := url.Parse(href); err == nil {
				if q := u.Query().Get("q"); q != "" && strings.HasPrefix(q, "http") {
					urls = append(urls, q)
				}
			}
		} else if strings.HasPrefix(href, "http") {
			urls = append(urls, href)
		}
	})
	return urls
}
