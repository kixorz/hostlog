package main

import (
	"log"
	"math"
	"time"
)

// VisibilityScore calculates a visibility score for a host based on its log events
// The score is calculated using the formula: α * e^(-λ * T) + β * V + γ * S
// Where:
// - T = Time since most recent event (in hours)
// - V = Event volume in the last hour
// - S = Severity score of recent events
// - α, β, γ = Weighting coefficients
// - λ = Decay rate constant
func VisibilityScore(host string) (float64, error) {
	// Default parameters
	alpha := 10.0 // Weight for time decay component
	beta := 0.5   // Weight for volume component
	gamma := 5.0  // Weight for severity component
	lambda := 0.2 // Decay rate constant

	// Calculate each component
	timeDecay, err := TimeDecayComponent(host, alpha, lambda)
	if err != nil {
		return 0, err
	}

	volume, err := VolumeComponent(host, beta)
	if err != nil {
		return 0, err
	}

	severity, err := SeverityComponent(host, gamma)
	if err != nil {
		return 0, err
	}

	// Calculate final score
	score := timeDecay + volume + severity

	// Ensure minimum visibility (optional)
	minScore := 0.1
	if score < minScore {
		score = minScore
	}

	return score, nil
}

// TimeDecayComponent calculates the time decay component of the visibility score
// Formula: α * e^(-λ * T)
// Where T is the time since the most recent event in hours
func TimeDecayComponent(host string, alpha, lambda float64) (float64, error) {
	// Get the most recent log for this host
	var log Log
	result := db.Where("client_ip = ?", host).Order("created_at desc").First(&log)
	if result.Error != nil {
		return 0, result.Error
	}

	// Calculate hours since the most recent event
	hoursSince := time.Since(log.Timestamp).Hours()

	// Calculate time decay component
	return alpha * math.Exp(-lambda*hoursSince), nil
}

// VolumeComponent calculates the volume component of the visibility score
// Formula: β * V
// Where V is the number of events in the last hour
func VolumeComponent(host string, beta float64) (float64, error) {
	// Calculate the timestamp for one hour ago
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// Count logs in the last hour
	var count int64
	result := db.Model(&Log{}).Where("client_ip = ? AND timestamp > ?", host, oneHourAgo).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	// Calculate volume component
	// Optional: Cap the volume to prevent unusual bursts from dominating
	maxVolume := 100.0
	volume := float64(count)
	if volume > maxVolume {
		volume = maxVolume
	}

	return beta * volume, nil
}

// SeverityComponent calculates the severity component of the visibility score
// Formula: γ * S
// Where S is a weighted average of event severities
func SeverityComponent(host string, gamma float64) (float64, error) {
	// Define time window for severity calculation (e.g., last 24 hours)
	timeWindow := time.Now().Add(-24 * time.Hour)

	// Get logs within the time window
	var logs []Log
	result := db.Where("client_ip = ? AND timestamp > ?", host, timeWindow).Find(&logs)
	if result.Error != nil {
		return 0, result.Error
	}

	if len(logs) == 0 {
		return 0, nil
	}

	// Define weights for different severity levels
	// Using the priority field from logs (lower number = higher severity in syslog)
	var errorCount, warningCount, infoCount int

	for _, log := range logs {
		// Extract severity from priority (lower 3 bits of priority)
		severity := log.Priority & 7

		switch severity {
		case 0, 1, 2: // Error levels
			errorCount++
		case 3, 4: // Warning levels
			warningCount++
		case 5, 6, 7: // Info and debug levels
			infoCount++
		}
	}

	// Calculate weighted severity score
	// Higher weights for more severe events
	errorWeight := 10.0
	warningWeight := 5.0
	infoWeight := 1.0

	totalCount := errorCount + warningCount + infoCount
	severityScore := (errorWeight*float64(errorCount) +
		warningWeight*float64(warningCount) +
		infoWeight*float64(infoCount)) / float64(totalCount)

	return gamma * severityScore, nil
}

func GetAllHosts() ([]string, error) {
	var hosts []string
	result := db.Model(&Log{}).Distinct("client_ip").Pluck("client_ip", &hosts)
	return hosts, result.Error
}

// GetAllHostScores calculates visibility scores for all hosts
// Returns a map of hostname to score
func GetAllHostScores() (map[string]float64, error) {
	hosts, err := GetAllHosts()
	if err != nil {
		return nil, err
	}

	scores := make(map[string]float64)
	for _, host := range hosts {
		if host == "" {
			continue // Skip empty hosts
		}

		score, err := VisibilityScore(host)
		if err != nil {
			// Log the error but continue with other hosts
			log.Printf("Error calculating score for host %s: %v", host, err)
			continue
		}

		scores[host] = score
	}

	return scores, nil
}
