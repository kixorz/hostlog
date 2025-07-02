package main

import (
	"embed"
	"fmt"
	"gopkg.in/mcuadros/go-syslog.v2"
	"hostlog/models"
	"log"
)

//go:embed static/* templates/*
var staticFiles embed.FS

func main() {
	_, err := models.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Set up syslog server
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)
	server.ListenUDP("0.0.0.0:514")
	err = server.Boot()
	if err != nil {
		log.Fatalf("Failed to start syslog server: %v", err)
	}

	fmt.Println("Syslog server started. Listening on UDP port 514...")

	// Process incoming log messages
	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			// Save log message to database
			if err := models.SaveLog(logParts); err != nil {
				log.Printf("Error saving log: %v", err)
			} else {
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
	SetupHTTP(staticFiles)
	go StartHTTPServer("8080")

	server.Wait()
}
