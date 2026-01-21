package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) (serverCmd *exec.Cmd) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "emacs-mcp-test")
	require.NoError(t, err)

	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "emacs-mcp"))
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build emacs-mcp binary: %s", output)
	serverCmd = exec.Command(filepath.Join(tmpDir, "emacs-mcp"))
	return serverCmd
}

// TestMCPServerIntegration tests the MCP server end-to-end
func TestMCPServerIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a client to connect to your server
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	serverCmd := buildBinary(t)
	transport := &mcp.CommandTransport{
		Command: serverCmd,
	}

	// Connect to the server
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer session.Close()

	t.Run("ListTools", func(t *testing.T) {
		result, err := session.ListTools(ctx, nil)
		if err != nil {
			t.Fatalf("ListTools failed: %v", err)
		}

		if len(result.Tools) == 0 {
			t.Error("Expected at least one tool, got none")
		}

		t.Logf("Found %d tools", len(result.Tools))
		for _, tool := range result.Tools {
			t.Logf("  - %s: %s", tool.Name, tool.Description)
		}
	})

	t.Run("CallTool", func(t *testing.T) {
		params := &mcp.CallToolParams{
			Name: "eval_lisp",
			Arguments: map[string]any{
				"code": "(message \"Hello, World!\")",
			},
		}

		result, err := session.CallTool(ctx, params)
		if err != nil {
			t.Fatalf("CallTool failed: %v", err)
		}

		if result.IsError {
			t.Fatalf("Tool returned error: %+v", result.Content[0])
		}

		assert.Equal(t, 1, len(result.Content))
		textContent, ok := result.Content[0].(*mcp.TextContent)
		assert.True(t, ok)
		assert.Contains(t, textContent.Text, "Hello, World!")
	})
}

// // TestMCPServerToolValidation tests tool input validation
// func TestMCPServerToolValidation(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	client := mcp.NewClient(&mcp.Implementation{
// 		Name:    "test-client",
// 		Version: "1.0.0",
// 	}, nil)

// 	transport := &mcp.CommandTransport{
// 		Command: exec.Command("your-server-command"),
// 	}

// 	session, err := client.Connect(ctx, transport, nil)
// 	if err != nil {
// 		t.Fatalf("Failed to connect to server: %v", err)
// 	}
// 	defer session.Close()

// 	t.Run("InvalidToolName", func(t *testing.T) {
// 		params := &mcp.CallToolParams{
// 			Name:      "nonexistent_tool",
// 			Arguments: map[string]any{},
// 		}

// 		_, err := session.CallTool(ctx, params)
// 		if err == nil {
// 			t.Error("Expected error for nonexistent tool, got none")
// 		}
// 	})

// 	t.Run("MissingRequiredArguments", func(t *testing.T) {
// 		params := &mcp.CallToolParams{
// 			Name:      "example_tool",
// 			Arguments: map[string]any{}, // Missing required args
// 		}

// 		result, err := session.CallTool(ctx, params)
// 		if err == nil && !result.IsError {
// 			t.Error("Expected error for missing required arguments")
// 		}
// 	})
// }

// // TestMCPServerConcurrency tests concurrent tool calls
// func TestMCPServerConcurrency(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
// 	defer cancel()

// 	client := mcp.NewClient(&mcp.Implementation{
// 		Name:    "test-client",
// 		Version: "1.0.0",
// 	}, nil)

// 	transport := &mcp.CommandTransport{
// 		Command: exec.Command("your-server-command"),
// 	}

// 	session, err := client.Connect(ctx, transport, nil)
// 	if err != nil {
// 		t.Fatalf("Failed to connect to server: %v", err)
// 	}
// 	defer session.Close()

// 	// Run multiple concurrent calls
// 	const numCalls = 10
// 	errChan := make(chan error, numCalls)

// 	for i := 0; i < numCalls; i++ {
// 		go func(idx int) {
// 			params := &mcp.CallToolParams{
// 				Name: "example_tool",
// 				Arguments: map[string]any{
// 					"param1": idx,
// 				},
// 			}

// 			result, err := session.CallTool(ctx, params)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}

// 			if result.IsError {
// 				errChan <- err
// 				return
// 			}

// 			errChan <- nil
// 		}(i)
// 	}

// 	// Collect results
// 	for i := 0; i < numCalls; i++ {
// 		if err := <-errChan; err != nil {
// 			t.Errorf("Concurrent call %d failed: %v", i, err)
// 		}
// 	}
// }

// // TestMCPServerContextCancellation tests context cancellation handling
// func TestMCPServerContextCancellation(t *testing.T) {
// 	ctx := context.Background()

// 	client := mcp.NewClient(&mcp.Implementation{
// 		Name:    "test-client",
// 		Version: "1.0.0",
// 	}, nil)

// 	transport := &mcp.CommandTransport{
// 		Command: exec.Command("your-server-command"),
// 	}

// 	session, err := client.Connect(ctx, transport, nil)
// 	if err != nil {
// 		t.Fatalf("Failed to connect to server: %v", err)
// 	}
// 	defer session.Close()

// 	// Create a context that will be cancelled
// 	callCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
// 	defer cancel()

// 	params := &mcp.CallToolParams{
// 		Name: "slow_tool", // Replace with a tool that takes time
// 		Arguments: map[string]any{
// 			"duration": "5s",
// 		},
// 	}

// 	_, err = session.CallTool(callCtx, params)
// 	if err == nil {
// 		t.Error("Expected error due to context cancellation")
// 	}

// 	if ctx.Err() != context.DeadlineExceeded {
// 		t.Logf("Got error: %v", err)
// 	}
// }

// // TestMCPServerResourceAccess tests resource reading
// func TestMCPServerResourceAccess(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	client := mcp.NewClient(&mcp.Implementation{
// 		Name:    "test-client",
// 		Version: "1.0.0",
// 	}, nil)

// 	transport := &mcp.CommandTransport{
// 		Command: exec.Command("your-server-command"),
// 	}

// 	session, err := client.Connect(ctx, transport, nil)
// 	if err != nil {
// 		t.Fatalf("Failed to connect to server: %v", err)
// 	}
// 	defer session.Close()

// 	t.Run("ReadResource", func(t *testing.T) {
// 		// First list resources to get a valid URI
// 		listResult, err := session.ListResources(ctx, nil)
// 		if err != nil {
// 			t.Fatalf("ListResources failed: %v", err)
// 		}

// 		if len(listResult.Resources) == 0 {
// 			t.Skip("No resources available to test")
// 		}

// 		// Try to read the first resource
// 		params := &mcp.ReadResourceParams{
// 			URI: listResult.Resources[0].URI,
// 		}

// 		result, err := session.ReadResource(ctx, params)
// 		if err != nil {
// 			t.Fatalf("ReadResource failed: %v", err)
// 		}

// 		if len(result.Contents) == 0 {
// 			t.Error("Expected content from resource")
// 		}

// 		for i, content := range result.Contents {
// 			if textContent, ok := content.(*mcp.TextResourceContents); ok {
// 				t.Logf("Resource content[%d]: %d bytes", i, len(textContent.Text))
// 			}
// 		}
// 	})
// }

// // BenchmarkMCPServerToolCall benchmarks tool call performance
// func BenchmarkMCPServerToolCall(b *testing.B) {
// 	ctx := context.Background()

// 	client := mcp.NewClient(&mcp.Implementation{
// 		Name:    "test-client",
// 		Version: "1.0.0",
// 	}, nil)

// 	transport := &mcp.CommandTransport{
// 		Command: exec.Command("your-server-command"),
// 	}

// 	session, err := client.Connect(ctx, transport, nil)
// 	if err != nil {
// 		b.Fatalf("Failed to connect to server: %v", err)
// 	}
// 	defer session.Close()

// 	params := &mcp.CallToolParams{
// 		Name: "example_tool",
// 		Arguments: map[string]any{
// 			"param1": "value1",
// 		},
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		result, err := session.CallTool(ctx, params)
// 		if err != nil {
// 			b.Fatalf("CallTool failed: %v", err)
// 		}
// 		if result.IsError {
// 			b.Fatalf("Tool returned error")
// 		}
// 	}
// }
