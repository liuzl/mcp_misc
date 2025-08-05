package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/liuzl/ai/gemini"
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
	client := gemini.NewClient(os.Getenv("GEMINI_API_KEY"), gemini.WithBaseURL(os.Getenv("GEMINI_BASE_URL")))

	fmt.Printf("%s%s🤖 Gemini MCP Agent Ready%s\n", ColorBold, ColorPurple, ColorReset)
	fmt.Printf("%sType 'exit' to quit%s\n\n", ColorGray, ColorReset)

	scanner := bufio.NewScanner(os.Stdin)
	var conversation []gemini.Content
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
		// Send message to Gemini
		fmt.Printf("%s%sGemini: %s", ColorBold, ColorGreen, ColorReset)

		Content := gemini.Content{
			Parts: []gemini.Part{{Text: gemini.StringPtr(userInput)}},
			Role:  gemini.StringPtr("user"),
		}
		conversation = append(conversation, Content)
		geminiResp, err := client.GenerateContent(context.Background(), "gemini-2.5-flash", &gemini.GenerateContentRequest{Contents: conversation})
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

		for _, candidate := range geminiResp.Candidates {
			for _, part := range candidate.Content.Parts {
				if *part.Text != "" {
					fmt.Printf("%s%s%s", ColorGreen, *part.Text, ColorReset)
				}
			}
			candidate.Content.Role = gemini.StringPtr("model")
			conversation = append(conversation, candidate.Content)
		}
		fmt.Println()
	}
}
