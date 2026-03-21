package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"自定义mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Add tool
	tool := mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)

	// Add tool handler
	s.AddTool(tool, helloHandler)

	// 创建 http 路由
	mux := http.NewServeMux()
	mux.HandleFunc("/management/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":{"http":{"working":true,"message":"conn ok."}}}`))
	})

	// 创建 StreamableHttp 服务器，并绑定自定义 httpServer
	srv := server.NewStreamableHTTPServer(
		s,
		server.WithStreamableHTTPServer(&http.Server{Handler: mux}),
	)
	// 使用自定义 httpServer 后，需要将显示指定 /mcp 路由
	mux.Handle("/mcp", srv)

	// 启动 mcp 服务
	if err := srv.Start(":8085"); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}
