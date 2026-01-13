package main

import (
	"embed"
	"flag"
	"fmt"
	"hostlog/models"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/mcuadros/go-syslog.v2"
)

//go:embed static/* templates/*
var staticFiles embed.FS

func main() {
	mcpFlag := flag.Bool("mcp", false, "Run as an MCP server")
	flag.Parse()

	if *mcpFlag {
		runMCPServer()
		return
	}

	_, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Set up syslog server
	syslogPort := os.Getenv("HOSTLOG_SYSLOG_PORT")
	if syslogPort == "" {
		syslogPort = "514"
	}

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)
	server.ListenTCP("0.0.0.0:" + syslogPort)
	server.ListenUDP("0.0.0.0:" + syslogPort)
	err = server.Boot()
	if err != nil {
		log.Fatalf("Failed to start syslog server: %v", err)
	}

	fmt.Printf("Syslog server started. Listening on port %s...\n", syslogPort)

	// Process incoming log messages
	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			// Save log message to database
			if logEntry, err := models.SaveLog(logParts); err != nil {
				log.Printf("Error saving log: %v", err)
			} else {
				// Send to SSE broadcaster
				logBroadcaster.Messages <- logEntry

				// Print a brief confirmation (optional)
				if content, ok := logParts["content"].(string); ok {
					fmt.Printf("Saved log: %s\n", content)
				} else {
					fmt.Println("Saved log entry")
				}
			}
		}
	}(channel)

	// Set up and start HTTP server
	httpPort := os.Getenv("HOSTLOG_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	go StartHTTPServer("8080", staticFiles)

	server.Wait()
}

func runMCPServer() {
	_, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	s := NewMCPServer()
	server.ServeStdio(s)
}
