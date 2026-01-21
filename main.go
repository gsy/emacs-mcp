package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"Emacs-Bridge",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery())

	toolInsertCode := mcp.NewTool("insert_code",
		mcp.WithDescription("Insert code into the active Emacs buffer"),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The code to insert"),
		),
	)

	// This tool wraps the 'emacsclient' command
	s.AddTool(toolInsertCode, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		// Run emacsclient to call our Lisp function
		cmd := exec.Command("emacsclient", "--eval",
			"(my/ai-insert-at-point \""+content+"\")")

		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(string(output)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
