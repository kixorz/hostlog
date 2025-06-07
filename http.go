package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
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
	http.HandleFunc("/messages", handleMessages)
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Create template functions map
	funcMap := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
	}

	// Parse templates with functions
	templates = template.Must(template.New("").Funcs(funcMap).ParseFS(staticFiles, "templates/*.html"))
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

	hostScores, err := GetAllHostScores()
	if err != nil {
		log.Printf("Error retrieving host scores: %v", err)
		hostScores = []HostScore{}
	}

	topHostScores := GetTopHostScores(hostScores, 3)

	// Prepare data for template
	data := struct {
		Logs     []LogDisplay
		Hosts    []HostScore
		TopHosts []HostScore
	}{
		Logs:     formatLogsForDisplay(logs),
		Hosts:    hostScores,
		TopHosts: topHostScores,
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

// handleMessages handles requests for the messages endpoint
func handleMessages(w http.ResponseWriter, r *http.Request) {
	hosts := r.URL.Query()["h"]
	page := 0
	if pageStr := r.URL.Query().Get("p"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	logs, err := GetFilteredLogs(hosts, page)
	if err != nil {
		log.Printf("Error retrieving logs: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Logs []LogDisplay
		Page int
	}{
		Logs: formatLogsForDisplay(logs),
		Page: page,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err = templates.ExecuteTemplate(w, "messages", data)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
