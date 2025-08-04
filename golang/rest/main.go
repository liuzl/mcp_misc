package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
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

type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	baseURL := os.Getenv("GEMINI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	model := "gemini-2.5-flash"
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, model, apiKey)
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
		// Send message to Gemini
		fmt.Printf("%s%sGemini: %s", ColorBold, ColorGreen, ColorReset)
		reqBody := GeminiRequest{
			Contents: []Content{
				{
					Parts: []Part{
						{Text: userInput},
					},
				},
			},
		}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("%sError: HTTP %d - %s%s\n", ColorRed, resp.StatusCode, string(body), ColorReset)
			continue
		}
		var geminiResp GeminiResponse
		if err := json.Unmarshal(body, &geminiResp); err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}
		for _, candidate := range geminiResp.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					fmt.Printf("%s%s%s", ColorGreen, part.Text, ColorReset)
				}
			}
		}
		fmt.Println()
	}
}
