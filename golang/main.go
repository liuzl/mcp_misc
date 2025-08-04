package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/genai"
)

// ANSI color codes
const (
	ColorBlue   = "\033[94m"
	ColorGreen  = "\033[92m"
	ColorYellow = "\033[93m"
	ColorRed    = "\033[91m"
	ColorPurple = "\033[95m"
	ColorCyan   = "\033[96m"
	ColorWhite  = "\033[97m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"
)

func main() {
	// Create Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      os.Getenv("GEMINI_API_KEY"),
		HTTPOptions: genai.HTTPOptions{BaseURL: os.Getenv("GEMINI_BASE_URL")},
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	fmt.Printf("%s%s🤖 Gemini MCP Agent Ready%s\n", ColorBold, ColorPurple, ColorReset)
	fmt.Printf("%sType 'exit' to quit%s\n\n", ColorGray, ColorReset)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("%s%sYou: %s", ColorBold, ColorBlue, ColorReset)

		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if strings.ToLower(userInput) == "exit" {
			fmt.Printf("\n%sGoodbye!%s\n", ColorGray, ColorReset)
			break
		}

		if userInput == "" {
			continue
		}

		// Send message to Gemini with streaming
		fmt.Printf("%s%sGemini: %s", ColorBold, ColorGreen, ColorReset)

		// Use gemini-2.5-flash model
		model := "gemini-2.5-flash"
		contents := genai.Text(userInput)

		seq := client.Models.GenerateContentStream(ctx, model, contents, nil)
		for resp, err := range seq {
			if err != nil {
				fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
				break
			}

			for _, candidate := range resp.Candidates {
				for _, part := range candidate.Content.Parts {
					if part.Text != "" {
						fmt.Printf("%s%s%s", ColorGreen, part.Text, ColorReset)
					}
				}
			}
		}
		fmt.Println()
	}
}
