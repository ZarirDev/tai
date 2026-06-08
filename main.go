package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
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
	return s.Err()
}

func main() {
	// Load .env
	if err := loadEnvFile(".env"); err != nil {
		log.Printf("Warning: .env load error: %v", err)
	}

	// Debug: show key prefix (safe)
	key := os.Getenv("GROQ_API_KEY")
	if key == "" {
		log.Fatal("GROQ_API_KEY not set. Use .env or export.")
	}
	if len(key) >= 4 {
		log.Printf("✅ API key loaded: %s...", key[:4])
	} else {
		log.Fatal("GROQ_API_KEY is too short.")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: tai <query> [URLs...]")
		os.Exit(1)
	}

	// Split args into URLs and words
	var urls []string
	var words []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			urls = append(urls, arg)
		} else {
			words = append(words, arg)
		}
	}
	userQuery := strings.Join(words, " ")

	// Case 1: explicit URLs provided
	if len(urls) > 0 {
		context, err := scrapeMultiple(urls)
		if err != nil {
			log.Fatalf("Scrape error: %v", err)
		}
		ans, err := askGroq(userQuery, context)
		if err != nil {
			log.Fatalf("Groq error: %v", err)
		}
		fmt.Println(ans)
		return
	}

	// Case 2: no URLs – decide search
	if userQuery == "" {
		fmt.Println("No query provided.")
		os.Exit(1)
	}

	needsSearch, err := requiresWebSearch(userQuery)
	if err != nil {
		log.Fatalf("Decision error: %v", err)
	}

	if needsSearch {
		fmt.Fprintln(os.Stderr, "🔍 Searching the web...")
		context, err := searchAndScrape(userQuery)
		if err != nil {
			log.Fatalf("Search failed: %v", err)
		}
		ans, err := askGroq(userQuery, context)
		if err != nil {
			log.Fatalf("Groq error: %v", err)
		}
		fmt.Println(ans)
	} else {
		ans, err := askGroq(userQuery, "")
		if err != nil {
			log.Fatalf("Groq error: %v", err)
		}
		fmt.Println(ans)
	}
}
