package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return nil // .env is optional
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

func main() {
	// Load .env file
	if err := loadEnvFile(".env"); err != nil {
		log.Printf("Warning: could not load .env: %v", err)
	}
	if os.Getenv("GROQ_API_KEY") == "" {
		log.Fatal("GROQ_API_KEY not set. Please create a .env file with GROQ_API_KEY=... or export it.")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: ai <query or URL> [more URLs...]")
		os.Exit(1)
	}

	args := os.Args[1:]

	var urls []string
	var queryWords []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			urls = append(urls, arg)
		} else {
			queryWords = append(queryWords, arg)
		}
	}
	userQuery := strings.Join(queryWords, " ")

	// Case 1: user provided URLs
	if len(urls) > 0 {
		contextText, err := scrapeMultiple(urls)
		if err != nil {
			log.Fatalf("Error scraping URLs: %v", err)
		}
		answer, err := askGroq(userQuery, contextText)
		if err != nil {
			log.Fatalf("Groq error: %v", err)
		}
		fmt.Println(answer)
		return
	}

	// Case 2: no URLs – decide if web search needed
	if userQuery == "" {
		fmt.Println("No query provided.")
		os.Exit(1)
	}

	needsSearch, err := requiresWebSearch(userQuery)
	if err != nil {
		log.Fatalf("Error checking if search is needed: %v", err)
	}

	if needsSearch {
		fmt.Fprintln(os.Stderr, "🔍 Searching the web...")
		contextText, err := searchAndScrape(userQuery)
		if err != nil {
			log.Fatalf("Web search failed: %v", err)
		}
		answer, err := askGroq(userQuery, contextText)
		if err != nil {
			log.Fatalf("Groq error: %v", err)
		}
		fmt.Println(answer)
	} else {
		answer, err := askGroq(userQuery, "")
		if err != nil {
			log.Fatalf("Groq error: %v", err)
		}
		fmt.Println(answer)
	}
}
