package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v25.8.0"}, nil)
	transport := mcp.NewStreamableClientTransport("http://localhost:8080/mcp", nil)
	session, err := client.Connect(ctx, transport)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, tool := range tools.Tools {
		log.Printf("Tool: %v, %v", tool.Name, tool.Description)
		if tool.Name == "get_os_info" {
			log.Printf("Calling tool: %v", tool.Name)
			params := &mcp.CallToolParams{
				Name:      "get_os_info",
				Arguments: map[string]any{},
			}
			res, err := session.CallTool(ctx, params)
			if err != nil {
				log.Fatal(err)
			}
			if res.IsError {
				log.Fatal("tool failed")
			}
			for _, c := range res.Content {
				log.Print(c.(*mcp.TextContent).Text)
			}
		}
	}
}
