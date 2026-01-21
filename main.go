package main

import (
	"context"
	"errors"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	Code string `json:"code" jsonschema:"The code to execute in the running Emacs sesssion"`
}

type Output struct {
	Result string `json:"result" jsonschema:"The execution result"`
}

func ToolGeneric(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, Output, error) {
	if len(input.Code) == 0 {
		return &mcp.CallToolResult{IsError: true}, Output{Result: ""}, errors.New("eval expression can't be empty")
	}
	cmd := exec.Command("emacsclient", "--eval", input.Code)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &mcp.CallToolResult{IsError: true}, Output{Result: string(output)}, err
	}
	return &mcp.CallToolResult{IsError: false}, Output{Result: string(output)}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "emacs", Version: "v1.0.0"}, nil)
	mcp.AddTool(server,
		&mcp.Tool{Name: "eval_lisp", Description: "Execute any Elisp command in the running Emacs session"}, ToolGeneric)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
