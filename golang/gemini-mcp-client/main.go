package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/liuzl/ai/gemini"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

// Tool represents a discovered MCP tool
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Global conversation history
var conversationHistory []gemini.Content

// Global tools registry
var discoveredTools []Tool

// printConversationHistory displays the current conversation history
func printConversationHistory() {
	if len(conversationHistory) == 0 {
		fmt.Printf("%sNo conversation history.%s\n", ColorGray, ColorReset)
		return
	}

	fmt.Printf("%s--- Conversation History (%d messages) ---%s\n", ColorBold, len(conversationHistory), ColorReset)

	userCount := 0
	modelCount := 0
	functionCallCount := 0
	functionResponseCount := 0

	for i, content := range conversationHistory {
		role := "unknown"
		if content.Role != nil {
			role = *content.Role
		}

		// Count different types of messages
		switch role {
		case "user":
			userCount++
		case "model":
			modelCount++
		}

		fmt.Printf("%s[%d] %s: ", ColorCyan, i+1, role)

		// Print text content
		hasContent := false
		for _, part := range content.Parts {
			if part.Text != nil && *part.Text != "" {
				text := *part.Text
				if len(text) > 100 {
					text = text[:100] + "..."
				}
				fmt.Printf("%s%s%s", ColorWhite, text, ColorReset)
				hasContent = true
			}
			if part.FunctionCall != nil {
				fmt.Printf("%s[Function Call: %s]%s", ColorYellow, part.FunctionCall.Name, ColorReset)
				functionCallCount++
				hasContent = true
			}
			if part.FunctionResponse != nil {
				fmt.Printf("%s[Function Response: %s]%s", ColorGreen, part.FunctionResponse.Name, ColorReset)
				functionResponseCount++
				hasContent = true
			}
		}

		if !hasContent {
			fmt.Printf("%s[No content]%s", ColorGray, ColorReset)
		}
		fmt.Println()
	}

	fmt.Printf("%s----------------------------------------%s\n", ColorBold, ColorReset)
	fmt.Printf("%sStatistics: %d user messages, %d model responses, %d function calls, %d function responses%s\n",
		ColorPurple, userCount, modelCount, functionCallCount, functionResponseCount, ColorReset)
}

// showConversationStats displays conversation statistics
func showConversationStats() {
	if len(conversationHistory) == 0 {
		fmt.Printf("%sNo conversation history.%s\n", ColorGray, ColorReset)
		return
	}

	userCount := 0
	modelCount := 0
	functionCallCount := 0
	functionResponseCount := 0

	for _, content := range conversationHistory {
		role := "unknown"
		if content.Role != nil {
			role = *content.Role
		}

		switch role {
		case "user":
			userCount++
		case "model":
			modelCount++
		}

		for _, part := range content.Parts {
			if part.FunctionCall != nil {
				functionCallCount++
			}
			if part.FunctionResponse != nil {
				functionResponseCount++
			}
		}
	}

	fmt.Printf("%s--- Conversation Statistics ---%s\n", ColorBold, ColorReset)
	fmt.Printf("%sTotal messages: %d%s\n", ColorCyan, len(conversationHistory), ColorReset)
	fmt.Printf("%sUser messages: %d%s\n", ColorBlue, userCount, ColorReset)
	fmt.Printf("%sModel responses: %d%s\n", ColorGreen, modelCount, ColorReset)
	fmt.Printf("%sFunction calls: %d%s\n", ColorYellow, functionCallCount, ColorReset)
	fmt.Printf("%sFunction responses: %d%s\n", ColorPurple, functionResponseCount, ColorReset)
	fmt.Printf("%s------------------------------%s\n", ColorBold, ColorReset)
}

// clearConversationHistory clears the conversation history
func clearConversationHistory() {
	conversationHistory = make([]gemini.Content, 0)
	fmt.Printf("%sConversation history cleared.%s\n", ColorGreen, ColorReset)
}

// initializeConversationHistory initializes the conversation with a system message
func initializeConversationHistory() {
	systemMessage := fmt.Sprintf(`You are a helpful AI assistant connected to multiple external systems via MCP tools.
The current date is %s.
You can use tools to help users with their requests.
Always be helpful and provide clear, accurate responses.`, time.Now().Format("2006-01-02"))

	conversationHistory = []gemini.Content{
		{
			Parts: []gemini.Part{{Text: gemini.StringPtr(systemMessage)}},
			Role:  gemini.StringPtr("user"),
		},
	}
}

// discoverAndRegisterCapabilities connects to MCP servers and registers their tools
func discoverAndRegisterCapabilities(session *mcp.ClientSession) error {
	fmt.Printf("%s🤖 Starting dynamic discovery of all MCP server capabilities...%s\n", ColorBold, ColorReset)

	ctx := context.Background()
	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}

	if len(tools.Tools) == 0 {
		fmt.Printf("%s⚠️ No tools found on any connected servers.%s\n", ColorYellow, ColorReset)
		return nil
	}

	for _, tool := range tools.Tools {
		discoveredTool := Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}
		discoveredTools = append(discoveredTools, discoveredTool)
		fmt.Printf("  %s✅ Discovered and registered tool: %s%s\n", ColorGreen, tool.Name, ColorReset)
	}

	fmt.Printf("%s✨ Capability discovery complete!%s\n", ColorGreen, ColorReset)
	return nil
}

// convertToGeminiTools converts MCP tools to Gemini function declarations
func convertToGeminiTools() []gemini.FunctionDeclaration {
	var functionDeclarations []gemini.FunctionDeclaration

	for _, tool := range discoveredTools {
		// Use a simple object schema for all tools
		geminiSchema := &gemini.Schema{
			Type: "object",
		}

		functionDecl := gemini.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  geminiSchema,
		}
		functionDeclarations = append(functionDeclarations, functionDecl)
	}

	return functionDeclarations
}

// callMCPTool calls a specific MCP tool and returns the result
func callMCPTool(session *mcp.ClientSession, toolName string, args map[string]any) (map[string]any, error) {
	ctx := context.Background()

	toolResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})

	if err != nil {
		return map[string]any{"error": fmt.Sprintf("Tool execution failed: %v", err)}, nil
	}

	if toolResult.IsError {
		errorText := ""
		for _, content := range toolResult.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				errorText = textContent.Text
				break
			}
		}
		return map[string]any{"error": errorText}, nil
	}

	// Extract text content from tool result
	var resultText string
	for _, content := range toolResult.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			resultText = textContent.Text
			break
		}
	}

	return map[string]any{"result": resultText}, nil
}

// agentLoop handles the conversation loop with proper function calling
func agentLoop(prompt string, geminiClient *gemini.Client, session *mcp.ClientSession) (*gemini.GenerateContentResponse, error) {
	ctx := context.Background()

	// Convert MCP tools to Gemini function declarations
	geminiTools := convertToGeminiTools()
	if len(geminiTools) > 0 {
		fmt.Printf("%sWarning: No tools available for function calling%s\n", ColorYellow, ColorReset)
	}

	// Add user message to conversation history
	userContent := gemini.Content{
		Parts: []gemini.Part{{Text: gemini.StringPtr(prompt)}},
		Role:  gemini.StringPtr("user"),
	}
	conversationHistory = append(conversationHistory, userContent)

	// Initial request with conversation history and function declarations
	request := &gemini.GenerateContentRequest{
		Contents: conversationHistory,
	}

	if len(geminiTools) > 0 {
		request.Tools = []gemini.Tool{
			{
				FunctionDeclarations: geminiTools,
			},
		}
	}

	response, err := geminiClient.GenerateContent(ctx, os.Getenv("GEMINI_MODEL"), request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate initial content: %v", err)
	}

	// Append initial response to conversation history
	if len(response.Candidates) > 0 {
		conversationHistory = append(conversationHistory, response.Candidates[0].Content)
	}

	// Tool calling loop
	turnCount := 0
	maxToolTurns := 30

	for turnCount < maxToolTurns {
		turnCount++

		// Check if response has function calls
		var functionCalls []gemini.FunctionCall
		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					functionCalls = append(functionCalls, *part.FunctionCall)
				}
			}
		}

		if len(functionCalls) == 0 {
			// No function calls, break the loop
			break
		}

		fmt.Printf("%sProcessing %d function call(s)...%s\n", ColorCyan, len(functionCalls), ColorReset)

		// Process all function calls in this turn
		var toolResponseParts []gemini.Part
		for _, functionCall := range functionCalls {
			toolName := functionCall.Name
			args := functionCall.Args
			if args == nil {
				args = make(map[string]interface{})
			}

			fmt.Printf("%sAttempting to call MCP tool: '%s' with args: %v%s\n", ColorCyan, toolName, args, ColorReset)

			toolResponse, err := callMCPTool(session, toolName, args)
			if err != nil {
				fmt.Printf("%sMCP tool '%s' execution failed: %v%s\n", ColorRed, toolName, err, ColorReset)
				toolResponse = map[string]any{"error": fmt.Sprintf("Tool execution failed: %v", err)}
			} else {
				fmt.Printf("%sMCP tool '%s' executed successfully%s\n", ColorGreen, toolName, ColorReset)
			}

			// Create function response part
			functionResponse := gemini.FunctionResponse{
				Name:     toolName,
				Response: toolResponse,
			}

			toolResponseParts = append(toolResponseParts, gemini.Part{
				FunctionResponse: &functionResponse,
			})
		}

		// Add tool response(s) to conversation history
		toolResponseContent := gemini.Content{
			Parts: toolResponseParts,
			Role:  gemini.StringPtr("user"),
		}
		conversationHistory = append(conversationHistory, toolResponseContent)

		fmt.Printf("%sAdded %d tool response parts to conversation history.%s\n", ColorCyan, len(toolResponseParts), ColorReset)

		// Make the next call to the model with updated conversation history
		fmt.Printf("%sMaking subsequent API call with tool responses...%s\n", ColorCyan, ColorReset)
		nextRequest := &gemini.GenerateContentRequest{
			Contents: conversationHistory,
		}
		if len(geminiTools) > 0 {
			nextRequest.Tools = []gemini.Tool{
				{
					FunctionDeclarations: geminiTools,
				},
			}
		}
		response, err = geminiClient.GenerateContent(ctx, os.Getenv("GEMINI_MODEL"), nextRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to generate content with tool result: %v", err)
		}

		// Append response to conversation history
		if len(response.Candidates) > 0 {
			conversationHistory = append(conversationHistory, response.Candidates[0].Content)
		}
	}

	if turnCount >= maxToolTurns {
		fmt.Printf("%sMaximum tool turns (%d) reached.%s\n", ColorYellow, maxToolTurns, ColorReset)
	}

	fmt.Printf("%sMCP tool calling loop finished.%s\n", ColorGreen, ColorReset)
	return response, nil
}

func main() {
	fmt.Printf("%s--- Gemini Universal MCP Client ---%s\n", ColorBold, ColorReset)

	// Initialize conversation history with system message
	initializeConversationHistory()

	// Initialize Gemini client
	geminiClient := gemini.NewClient(os.Getenv("GEMINI_API_KEY"), gemini.WithBaseURL(os.Getenv("GEMINI_BASE_URL")))

	// Initialize MCP client and session
	ctx := context.Background()
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "gemini-mcp-client", Version: "v1.0.0"}, nil)

	// Try to connect to MCP server
	var session *mcp.ClientSession
	var err error

	// Get MCP server URL from environment or use default
	mcpServerURL := os.Getenv("MCP_SERVER_URL")
	if mcpServerURL == "" {
		mcpServerURL = "http://localhost:8080/mcp"
	}

	fmt.Printf("%sConnecting to MCP server: %s%s\n", ColorCyan, mcpServerURL, ColorReset)

	transport := mcp.NewStreamableClientTransport(mcpServerURL, nil)
	session, err = mcpClient.Connect(ctx, transport)
	if err != nil {
		fmt.Printf("%sWarning: Failed to connect to MCP server: %v%s\n", ColorYellow, err, ColorReset)
		fmt.Printf("%sContinuing without MCP tools...%s\n", ColorYellow, ColorReset)
		session = nil
	} else {
		defer session.Close()
		fmt.Printf("%s✅ Successfully connected to MCP server%s\n", ColorGreen, ColorReset)
	}

	// Discover capabilities and configure the LLM
	if session != nil {
		if err := discoverAndRegisterCapabilities(session); err != nil {
			fmt.Printf("%sWarning: Failed to discover capabilities: %v%s\n", ColorYellow, err, ColorReset)
		}
	}

	// Display discovered tools
	if len(discoveredTools) > 0 {
		fmt.Printf("\n--- Discovered Tools ---\n")
		for _, tool := range discoveredTools {
			fmt.Printf("%s- %s: %s%s\n", ColorGreen, tool.Name, tool.Description, ColorReset)
		}
		fmt.Printf("------------------------\n\n")
	} else {
		fmt.Printf("%sNo MCP tools available. Running in basic chat mode.%s\n", ColorYellow, ColorReset)
	}

	// Start interactive chat
	fmt.Printf("%s🤖 Universal MCP Agent Ready. Type 'exit' to quit.%s\n", ColorBold, ColorReset)
	fmt.Printf("%sCommands: 'exit', 'history', 'clear', 'stats'%s\n\n", ColorGray, ColorReset)

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

		// Handle special commands
		switch strings.ToLower(userInput) {
		case "history":
			printConversationHistory()
			continue
		case "clear":
			clearConversationHistory()
			continue
		case "stats":
			showConversationStats()
			continue
		}

		// Run agent loop
		fmt.Printf("%s%sGemini: %s", ColorBold, ColorGreen, ColorReset)

		response, err := agentLoop(userInput, geminiClient, session)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

		// Print response
		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				if part.Text != nil && *part.Text != "" {
					fmt.Printf("%s%s%s", ColorGreen, *part.Text, ColorReset)
				}
			}
		}
		fmt.Println()
	}
}
