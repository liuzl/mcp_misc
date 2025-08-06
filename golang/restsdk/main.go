package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/liuzl/ai"
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
	provider := os.Getenv("AI_PROVIDER")
	var apiKey, baseURL string
	switch provider {
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
		baseURL = os.Getenv("OPENAI_BASE_URL")
	case "gemini":
		apiKey = os.Getenv("GEMINI_API_KEY")
		baseURL = os.Getenv("GEMINI_BASE_URL")
	default:
		fmt.Printf("Unsupported AI_PROVIDER: %s", provider)
		return
	}

	if apiKey == "" {
		fmt.Printf("API key for %s not set, skipping test", provider)
		return
	}

	client, err := ai.NewClient(ai.WithProvider(provider), ai.WithAPIKey(apiKey), ai.WithBaseURL(baseURL))
	if err != nil {
		fmt.Printf("Error creating client: %v", err)
		return
	}

	fmt.Printf("%s%s🤖 Gemini MCP Agent Ready%s\n", ColorBold, ColorPurple, ColorReset)
	fmt.Printf("%sType 'exit' to quit%s\n\n", ColorGray, ColorReset)

	scanner := bufio.NewScanner(os.Stdin)
	var conversation []ai.Message
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

		Content := ai.Message{Role: "user", Content: userInput}
		conversation = append(conversation, Content)
		resp, err := client.Generate(context.Background(), &ai.Request{Model: "gemini-2.5-flash", Messages: conversation})
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}
		if resp.Text == "" {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

		fmt.Printf("%s%s%s", ColorGreen, resp.Text, ColorReset)
		fmt.Println()
	}
}
