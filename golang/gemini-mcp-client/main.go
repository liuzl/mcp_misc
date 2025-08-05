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

// conversationStats holds statistics about the conversation.
type conversationStats struct {
	UserMessages      int
	ModelResponses    int
	FunctionCalls     int
	FunctionResponses int
	TotalMessages     int
}

// Agent holds the state for a chat session, including conversation history and tools.
type Agent struct {
	geminiClient        *gemini.Client
	mcpSession          *mcp.ClientSession
	conversationHistory []gemini.Content
	discoveredTools     []Tool
}

// NewAgent creates and initializes a new Agent.
func NewAgent(geminiClient *gemini.Client, mcpSession *mcp.ClientSession) *Agent {
	agent := &Agent{
		geminiClient:    geminiClient,
		mcpSession:      mcpSession,
		discoveredTools: []Tool{},
	}
	agent.initializeConversation()
	return agent
}

// initializeConversation sets up the initial system message.
func (a *Agent) initializeConversation() {
	systemMessage := fmt.Sprintf(`You are a helpful AI assistant connected to multiple external systems via MCP tools.
The current date is %s.
You can use tools to help users with their requests.
Always be helpful and provide clear, accurate responses.`, time.Now().Format("2006-01-02"))

	a.conversationHistory = []gemini.Content{
		{
			Parts: []gemini.Part{{Text: gemini.StringPtr(systemMessage)}},
			Role:  gemini.StringPtr("user"),
		},
	}
}

// getStats calculates and returns statistics about the conversation.
func (a *Agent) getStats() conversationStats {
	stats := conversationStats{TotalMessages: len(a.conversationHistory)}
	for _, content := range a.conversationHistory {
		role := "unknown"
		if content.Role != nil {
			role = *content.Role
		}
		switch role {
		case "user":
			stats.UserMessages++
		case "model":
			stats.ModelResponses++
		}
		for _, part := range content.Parts {
			if part.FunctionCall != nil {
				stats.FunctionCalls++
			}
			if part.FunctionResponse != nil {
				stats.FunctionResponses++
			}
		}
	}
	return stats
}

// printConversationHistory displays the current conversation history.
func (a *Agent) printConversationHistory() {
	if len(a.conversationHistory) == 0 {
		fmt.Printf("%sNo conversation history.%s\n", ColorGray, ColorReset)
		return
	}

	stats := a.getStats()
	fmt.Printf("%s--- Conversation History (%d messages) ---%s\n", ColorBold, stats.TotalMessages, ColorReset)

	for i, content := range a.conversationHistory {
		role := "unknown"
		if content.Role != nil {
			role = *content.Role
		}

		fmt.Printf("%s[%d] %s: ", ColorCyan, i+1, role)

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
				hasContent = true
			}
			if part.FunctionResponse != nil {
				fmt.Printf("%s[Function Response: %s]%s", ColorGreen, part.FunctionResponse.Name, ColorReset)
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
		ColorPurple, stats.UserMessages, stats.ModelResponses, stats.FunctionCalls, stats.FunctionResponses, ColorReset)
}

// showConversationStats displays conversation statistics.
func (a *Agent) showConversationStats() {
	if len(a.conversationHistory) == 0 {
		fmt.Printf("%sNo conversation history.%s\n", ColorGray, ColorReset)
		return
	}
	stats := a.getStats()
	fmt.Printf("%s--- Conversation Statistics ---%s\n", ColorBold, ColorReset)
	fmt.Printf("%sTotal messages: %d%s\n", ColorCyan, stats.TotalMessages, ColorReset)
	fmt.Printf("%sUser messages: %d%s\n", ColorBlue, stats.UserMessages, ColorReset)
	fmt.Printf("%sModel responses: %d%s\n", ColorGreen, stats.ModelResponses, ColorReset)
	fmt.Printf("%sFunction calls: %d%s\n", ColorYellow, stats.FunctionCalls, ColorReset)
	fmt.Printf("%sFunction responses: %d%s\n", ColorPurple, stats.FunctionResponses, ColorReset)
	fmt.Printf("%s------------------------------%s\n", ColorBold, ColorReset)
}

// clearConversationHistory clears the conversation history.
func (a *Agent) clearConversation() {
	a.initializeConversation()
	fmt.Printf("%sConversation history cleared.%s\n", ColorGreen, ColorReset)
}

// discoverTools connects to MCP servers and registers their tools.
func (a *Agent) discoverTools() error {
	if a.mcpSession == nil {
		return nil
	}
	fmt.Printf("%s🤖 Starting dynamic discovery of all MCP server capabilities...%s\n", ColorBold, ColorReset)

	ctx := context.Background()
	tools, err := a.mcpSession.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}

	if len(tools.Tools) == 0 {
		fmt.Printf("%s⚠️ No tools found on any connected servers.%s\n", ColorYellow, ColorReset)
		return nil
	}

	for _, tool := range tools.Tools {
		a.discoveredTools = append(a.discoveredTools, Tool{
			Name:        tool.Name,
			Description: tool.Description,
		})
		fmt.Printf("  %s✅ Discovered and registered tool: %s%s\n", ColorGreen, tool.Name, ColorReset)
	}

	fmt.Printf("%s✨ Capability discovery complete!%s\n", ColorGreen, ColorReset)
	return nil
}

// convertToGeminiTools converts MCP tools to Gemini function declarations.
func (a *Agent) convertToGeminiTools() []gemini.FunctionDeclaration {
	var functionDeclarations []gemini.FunctionDeclaration
	for _, tool := range a.discoveredTools {
		functionDecl := gemini.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  &gemini.Schema{Type: "object"}, // Use a simple object schema for all tools
		}
		functionDeclarations = append(functionDeclarations, functionDecl)
	}
	return functionDeclarations
}

// callMCPTool calls a specific MCP tool and returns the result.
func (a *Agent) callMCPTool(toolName string, args map[string]any) (map[string]any, error) {
	ctx := context.Background()
	toolResult, err := a.mcpSession.CallTool(ctx, &mcp.CallToolParams{
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

	var resultText string
	for _, content := range toolResult.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			resultText = textContent.Text
			break
		}
	}
	return map[string]any{"result": resultText}, nil
}

// agentLoop handles the conversation loop, including function calling.
func (a *Agent) agentLoop(prompt string) (*gemini.GenerateContentResponse, error) {
	ctx := context.Background()
	geminiTools := a.convertToGeminiTools()

	// Add user message to conversation history
	userContent := gemini.Content{
		Parts: []gemini.Part{{Text: gemini.StringPtr(prompt)}},
		Role:  gemini.StringPtr("user"),
	}
	a.conversationHistory = append(a.conversationHistory, userContent)

	// Initial request
	request := &gemini.GenerateContentRequest{Contents: a.conversationHistory}
	if len(geminiTools) > 0 {
		request.Tools = []gemini.Tool{{FunctionDeclarations: geminiTools}}
	}

	response, err := a.geminiClient.GenerateContent(ctx, os.Getenv("GEMINI_MODEL"), request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate initial content: %v", err)
	}

	if len(response.Candidates) > 0 {
		a.conversationHistory = append(a.conversationHistory, response.Candidates[0].Content)
	}

	// Tool calling loop
	maxToolTurns := 30
	for range maxToolTurns {
		var functionCalls []gemini.FunctionCall
		if len(response.Candidates) > 0 {
			for _, part := range response.Candidates[0].Content.Parts {
				if part.FunctionCall != nil {
					functionCalls = append(functionCalls, *part.FunctionCall)
				}
			}
		}

		if len(functionCalls) == 0 {
			break // No more function calls, exit loop
		}

		fmt.Printf("%sProcessing %d function call(s)...%s\n", ColorCyan, len(functionCalls), ColorReset)

		var toolResponseParts []gemini.Part
		for _, fc := range functionCalls {
			args := fc.Args
			if args == nil {
				args = make(map[string]any)
			}
			fmt.Printf("%sAttempting to call MCP tool: '%s' with args: %v%s\n", ColorCyan, fc.Name, args, ColorReset)

			toolResponse, err := a.callMCPTool(fc.Name, args)
			if err != nil {
				fmt.Printf("%sMCP tool '%s' execution failed: %v%s\n", ColorRed, fc.Name, err, ColorReset)
				toolResponse = map[string]any{"error": fmt.Sprintf("Tool execution failed: %v", err)}
			} else {
				fmt.Printf("%sMCP tool '%s' executed successfully%s\n", ColorGreen, fc.Name, ColorReset)
			}
			toolResponseParts = append(toolResponseParts, gemini.Part{
				FunctionResponse: &gemini.FunctionResponse{
					Name:     fc.Name,
					Response: toolResponse,
				},
			})
		}
		// Add tool response to history with the correct 'tool' role
		a.conversationHistory = append(a.conversationHistory, gemini.Content{
			Parts: toolResponseParts,
			Role:  gemini.StringPtr("tool"),
		})
		// Make the next call to the model
		nextRequest := &gemini.GenerateContentRequest{Contents: a.conversationHistory}
		if len(geminiTools) > 0 {
			nextRequest.Tools = []gemini.Tool{{FunctionDeclarations: geminiTools}}
		}
		response, err = a.geminiClient.GenerateContent(ctx, os.Getenv("GEMINI_MODEL"), nextRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to generate content with tool result: %v", err)
		}
		if len(response.Candidates) > 0 {
			a.conversationHistory = append(a.conversationHistory, response.Candidates[0].Content)
		}
	}

	fmt.Printf("%sMCP tool calling loop finished.%s\n", ColorGreen, ColorReset)
	return response, nil
}

// runChatLoop starts the interactive read-eval-print loop.
func runChatLoop(agent *Agent) {
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
		switch strings.ToLower(userInput) {
		case "history":
			agent.printConversationHistory()
			continue
		case "clear":
			agent.clearConversation()
			continue
		case "stats":
			agent.showConversationStats()
			continue
		}

		fmt.Printf("%s%sGemini: %s", ColorBold, ColorGreen, ColorReset)
		response, err := agent.agentLoop(userInput)
		if err != nil {
			fmt.Printf("%sError: %v%s\n", ColorRed, err, ColorReset)
			continue
		}

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

func main() {
	fmt.Printf("%s--- Gemini Universal MCP Client ---%s\n", ColorBold, ColorReset)

	// Initialize Gemini client
	geminiClient := gemini.NewClient(os.Getenv("GEMINI_API_KEY"), gemini.WithBaseURL(os.Getenv("GEMINI_BASE_URL")))

	// Initialize MCP client and session
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "gemini-mcp-client", Version: "v1.0.0"}, nil)
	mcpServerURL := os.Getenv("MCP_SERVER_URL")
	if mcpServerURL == "" {
		mcpServerURL = "http://localhost:8080/mcp"
	}

	fmt.Printf("%sConnecting to MCP server: %s%s\n", ColorCyan, mcpServerURL, ColorReset)
	transport := mcp.NewStreamableClientTransport(mcpServerURL, nil)
	session, err := mcpClient.Connect(context.Background(), transport)
	if err != nil {
		fmt.Printf("%sWarning: Failed to connect to MCP server: %v%s\n", ColorYellow, err, ColorReset)
		fmt.Printf("%sContinuing without MCP tools...%s\n", ColorYellow, ColorReset)
	} else {
		defer session.Close()
		fmt.Printf("%s✅ Successfully connected to MCP server%s\n", ColorGreen, ColorReset)
	}

	// Create and configure the agent
	agent := NewAgent(geminiClient, session)
	if err := agent.discoverTools(); err != nil {
		fmt.Printf("%sWarning: Failed to discover capabilities: %v%s\n", ColorYellow, err, ColorReset)
	}

	if len(agent.discoveredTools) > 0 {
		fmt.Printf("\n--- Discovered Tools ---\n")
		for _, tool := range agent.discoveredTools {
			fmt.Printf("%s- %s: %s%s\n", ColorGreen, tool.Name, tool.Description, ColorReset)
		}
		fmt.Printf("------------------------\n\n")
	} else {
		fmt.Printf("%sNo MCP tools available. Running in basic chat mode.%s\n", ColorYellow, ColorReset)
	}

	// Start interactive chat
	runChatLoop(agent)
}
