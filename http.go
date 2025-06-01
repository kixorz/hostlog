package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

// Templates for HTML rendering
var templates *template.Template

// SetupHTTP initializes the HTTP server with static files and dynamic routes
func SetupHTTP(staticFiles embed.FS) {
	// Set up static file serving
	staticContent, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("Failed to get static files: %v", err)
	}

	// Create file server handler for static files
	fileServer := http.FileServer(http.FS(staticContent))

	// Set up routes
	http.HandleFunc("/", handleIndex)
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Parse templates
	templates = template.Must(template.ParseFS(staticFiles, "templates/*.html"))
}

// StartHTTPServer starts the HTTP server on the specified port
func StartHTTPServer(port string) {
	log.Printf("Web server started. Listening on HTTP port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// handleIndex handles the main page request
func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Get recent logs from database
	logs, err := GetRecentLogs(100) // Get the 100 most recent logs
	if err != nil {
		log.Printf("Error retrieving logs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := struct {
		Logs []LogDisplay
	}{
		Logs: formatLogsForDisplay(logs),
	}

	// Render template
	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LogDisplay represents a log entry formatted for display
type LogDisplay struct {
	Timestamp string
	Source    string
	Severity  string
	Message   string
	Class     string // CSS class for styling based on severity
}

// formatLogsForDisplay converts database logs to display format
func formatLogsForDisplay(logs []Log) []LogDisplay {
	var displayLogs []LogDisplay

	for _, log := range logs {
		severity, class := getSeverityInfo(log.Priority)

		displayLog := LogDisplay{
			Timestamp: log.Timestamp.Format("2006-01-02 15:04:05"),
			Source:    log.ClientIP,
			Severity:  severity,
			Message:   log.Content,
			Class:     class,
		}

		displayLogs = append(displayLogs, displayLog)
	}

	return displayLogs
}

// getSeverityInfo returns the severity text and CSS class based on priority
func getSeverityInfo(priority int) (string, string) {
	// Syslog severity levels (0-7)
	switch priority & 7 {
	case 0, 1, 2:
		return "Error", "severity-error"
	case 3, 4:
		return "Warning", "severity-warning"
	case 5:
		return "Info", "severity-info"
	case 6, 7:
		return "Debug", "severity-debug"
	default:
		return "Unknown", ""
	}
}
