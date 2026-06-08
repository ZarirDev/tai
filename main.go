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
	// Load .env if present
	if err := loadEnvFile(".env"); err != nil {
		log.Printf("Warning: .env load error: %v", err)
	}

	// Validate API key
	key := os.Getenv("GROQ_API_KEY")
	if key == "" {
		log.Fatal("GROQ_API_KEY not set. Create a .env file or export it.")
	}
	if len(key) >= 4 {
		log.Printf("✓ API key loaded: %s...", key[:4])
	} else {
		log.Fatal("GROQ_API_KEY is too short.")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: tai <your question>")
		fmt.Println("Example: tai \"What is the weather in Paris?\"")
		os.Exit(1)
	}

	// Join all arguments into one query (supports spaces and quotes)
	query := strings.Join(os.Args[1:], " ")
	log.Printf("➡️  Asking: %s", query)

	answer, err := Ask(query)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	fmt.Println("\n🤖 Answer:")
	fmt.Println(answer)
}
