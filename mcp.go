package main

import (
	"context"
	"fmt"
	"hostlog/models"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func NewMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"hostlog",
		"1.0.0",
		server.WithLogging(),
	)

	// Tool to list all hosts
	s.AddTool(mcp.NewTool("list_hosts",
		mcp.WithDescription("List all hosts that have sent logs"),
	), listHostsHandler)

	// Tool to get logs for a specific host
	s.AddTool(mcp.NewTool("get_logs",
		mcp.WithDescription("Get recent logs, optionally filtered by host"),
		mcp.WithArray("hosts", mcp.Description("Optional list of host IPs to filter by"), mcp.WithStringItems()),
		mcp.WithNumber("page", mcp.Description("Page number (100 logs per page)"), mcp.DefaultNumber(0)),
	), getLogsHandler)

	// Tool to get host visibility scores
	s.AddTool(mcp.NewTool("get_host_scores",
		mcp.WithDescription("Get visibility scores for all hosts"),
	), getHostScoresHandler)

	return s
}

func listHostsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hosts, err := models.GetAllHosts()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get hosts: %v", err)), nil
	}

	text := "Hosts:\n"
	for _, host := range hosts {
		text += fmt.Sprintf("- %s\n", host)
	}

	return mcp.NewToolResultText(text), nil
}

func getLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var hosts []string
	// request.Params.Arguments is any, usually map[string]interface{}
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		// If no arguments provided or wrong format
		args = make(map[string]interface{})
	}

	if h, ok := args["hosts"]; ok {
		if slice, ok := h.([]interface{}); ok {
			for _, v := range slice {
				if s, ok := v.(string); ok {
					hosts = append(hosts, s)
				}
			}
		}
	}

	page := 0
	if p, ok := args["page"]; ok {
		if f, ok := p.(float64); ok {
			page = int(f)
		}
	}

	logs, _, err := models.GetFilteredLogs(hosts, page)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get logs: %v", err)), nil
	}

	text := fmt.Sprintf("Logs (Page %d):\n", page)
	for _, l := range logs {
		severity, _ := getSeverityInfo(l.Priority)
		text += fmt.Sprintf("[%s] %s [%s]: %s\n",
			l.Timestamp.Format("2006-01-02 15:04:05"),
			l.ClientIP,
			severity,
			l.Content)
	}

	if len(logs) == 0 {
		text += "No logs found."
	}

	return mcp.NewToolResultText(text), nil
}

func getHostScoresHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hostScores, err := GetAllHostScores()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get host scores: %v", err)), nil
	}

	text := "Host Visibility Scores:\n"
	for _, hs := range hostScores {
		text += fmt.Sprintf("- %s: %.2f\n", hs.Host, hs.Score)
	}

	if len(hostScores) == 0 {
		text += "No host scores found."
	}

	return mcp.NewToolResultText(text), nil
}

func ServeMCP() {
	s := NewMCPServer()
	// Use stdio transport
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
