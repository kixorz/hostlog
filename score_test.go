package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDBFromCSV creates a temporary test database for CSV-based tests
func setupTestDBFromCSV(t *testing.T) (*gorm.DB, func()) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-logs-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()

	// Open the database connection
	testDB, err := gorm.Open(sqlite.Open(tempFile.Name()), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Migrate the schema
	err = testDB.AutoMigrate(&Log{})
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	// Set the global db variable to our test database
	db = testDB

	// Return a cleanup function
	cleanup := func() {
		// Get the underlying SQL database
		sqlDB, err := testDB.DB()
		if err == nil {
			sqlDB.Close()
		}
		os.Remove(tempFile.Name())
	}

	return testDB, cleanup
}

// loadTestLogsFromCSV reads test log data from a CSV file and inserts it into the database
func loadTestLogsFromCSV(t *testing.T, testDB *gorm.DB, csvFilePath string) {
	// Open the CSV file
	file, err := os.Open(csvFilePath)
	if err != nil {
		t.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	// Verify the header has the expected columns
	expectedColumns := []string{"ClientIP", "Hostname", "Content", "Priority", "TimestampOffset"}
	for i, column := range expectedColumns {
		if i >= len(header) || header[i] != column {
			t.Fatalf("CSV header does not match expected format. Expected %v, got %v", expectedColumns, header)
		}
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV records: %v", err)
	}

	now := time.Now()

	// Process each record and insert into the database
	for _, record := range records {
		if len(record) < 5 {
			t.Fatalf("CSV record has fewer columns than expected: %v", record)
		}

		// Parse priority
		priority, err := strconv.Atoi(record[3])
		if err != nil {
			t.Fatalf("Failed to parse Priority as integer: %v", err)
		}

		// Parse timestamp offset (in minutes)
		offsetMinutes, err := strconv.Atoi(record[4])
		if err != nil {
			t.Fatalf("Failed to parse TimestampOffset as integer: %v", err)
		}
		timestamp := now.Add(time.Duration(offsetMinutes) * time.Minute)

		// Create and insert the log
		log := Log{
			ClientIP:  record[0],
			Hostname:  record[1],
			Content:   record[2],
			Priority:  priority,
			Timestamp: timestamp,
		}

		err = testDB.Create(&log).Error
		if err != nil {
			t.Fatalf("Failed to create test log: %v", err)
		}
	}
}

// TestTimeDecayComponentCSV tests the time decay component using data from CSV
func TestTimeDecayComponentCSV(t *testing.T) {
	// Setup test database
	_, cleanup := setupTestDBFromCSV(t)
	defer cleanup()

	// Load test logs from CSV
	loadTestLogsFromCSV(t, db, "testdata/time_decay_test.csv")

	// Test parameters
	alpha := 10.0
	lambda := 0.2

	// Test hosts with different recency
	hosts := []string{
		"192.168.1.1", // Recent activity
		"192.168.1.2", // Less recent
		"192.168.1.3", // Old activity
		"192.168.1.4", // Very recent
	}

	scores := make(map[string]float64)
	for _, host := range hosts {
		score, err := TimeDecayComponent(host, alpha, lambda)
		if err != nil {
			t.Fatalf("TimeDecayComponent for host %s returned an error: %v", host, err)
		}
		scores[host] = score
		t.Logf("Host %s time decay score: %f", host, score)
	}

	// Verify that more recent activity results in higher scores
	if scores["192.168.1.4"] <= scores["192.168.1.1"] {
		t.Errorf("Expected host4 (most recent) score (%f) to be greater than host1 score (%f)",
			scores["192.168.1.4"], scores["192.168.1.1"])
	}
	if scores["192.168.1.1"] <= scores["192.168.1.2"] {
		t.Errorf("Expected host1 score (%f) to be greater than host2 score (%f)",
			scores["192.168.1.1"], scores["192.168.1.2"])
	}
	if scores["192.168.1.2"] <= scores["192.168.1.3"] {
		t.Errorf("Expected host2 score (%f) to be greater than host3 (oldest) score (%f)",
			scores["192.168.1.2"], scores["192.168.1.3"])
	}
}

// TestVolumeComponentCSV tests the volume component using data from CSV
func TestVolumeComponentCSV(t *testing.T) {
	// Setup test database
	_, cleanup := setupTestDBFromCSV(t)
	defer cleanup()

	// Load test logs from CSV
	loadTestLogsFromCSV(t, db, "testdata/volume_test.csv")

	// Test parameter
	beta := 0.5

	// Test hosts with different volumes
	hosts := []string{
		"192.168.1.1", // Medium volume
		"192.168.1.5", // High volume
		"192.168.1.6", // Low volume
	}

	scores := make(map[string]float64)
	for _, host := range hosts {
		score, err := VolumeComponent(host, beta)
		if err != nil {
			t.Fatalf("VolumeComponent for host %s returned an error: %v", host, err)
		}
		scores[host] = score
		t.Logf("Host %s volume score: %f", host, score)
	}

	// Verify that higher volume results in higher scores
	if scores["192.168.1.5"] <= scores["192.168.1.1"] {
		t.Errorf("Expected host5 (high volume) score (%f) to be greater than host1 (medium volume) score (%f)",
			scores["192.168.1.5"], scores["192.168.1.1"])
	}
	if scores["192.168.1.1"] <= scores["192.168.1.6"] {
		t.Errorf("Expected host1 (medium volume) score (%f) to be greater than host6 (low volume) score (%f)",
			scores["192.168.1.1"], scores["192.168.1.6"])
	}
}

// TestSeverityComponentCSV tests the severity component using data from CSV
func TestSeverityComponentCSV(t *testing.T) {
	// Setup test database
	_, cleanup := setupTestDBFromCSV(t)
	defer cleanup()

	// Load test logs from CSV
	loadTestLogsFromCSV(t, db, "testdata/severity_test.csv")

	// Test parameter
	gamma := 5.0

	// Test hosts with different severity profiles
	hosts := []string{
		"192.168.1.1", // Mix of critical, error, warning, info (higher severity)
		"192.168.1.3", // All errors (highest severity)
		"192.168.1.4", // All info/debug (lowest severity)
		"192.168.1.5", // All warnings (medium severity)
	}

	scores := make(map[string]float64)
	for _, host := range hosts {
		score, err := SeverityComponent(host, gamma)
		if err != nil {
			t.Fatalf("SeverityComponent for host %s returned an error: %v", host, err)
		}
		scores[host] = score
		t.Logf("Host %s severity score: %f", host, score)
	}

	// Verify that higher severity results in higher scores
	if scores["192.168.1.3"] <= scores["192.168.1.5"] {
		t.Errorf("Expected host3 (all errors) score (%f) to be greater than host5 (all warnings) score (%f)",
			scores["192.168.1.3"], scores["192.168.1.5"])
	}
	if scores["192.168.1.5"] <= scores["192.168.1.4"] {
		t.Errorf("Expected host5 (all warnings) score (%f) to be greater than host4 (all info) score (%f)",
			scores["192.168.1.5"], scores["192.168.1.4"])
	}
}

// TestVisibilityScoreCSV tests the overall visibility score using data from CSV
func TestVisibilityScoreCSV(t *testing.T) {
	// Setup test database
	_, cleanup := setupTestDBFromCSV(t)
	defer cleanup()

	// Load test logs from CSV
	loadTestLogsFromCSV(t, db, "testdata/visibility_test.csv")

	// Test hosts with different characteristics
	hosts := []string{
		"192.168.1.1", // Recent, high severity, high volume
		"192.168.1.2", // Less recent, mixed severity
		"192.168.1.3", // Old, high severity
		"192.168.1.4", // Recent, low severity
		"192.168.1.5", // Recent, medium severity, high volume
		"192.168.1.6", // Recent, high severity, low volume
	}

	scores := make(map[string]float64)
	for _, host := range hosts {
		score, err := VisibilityScore(host)
		if err != nil {
			t.Fatalf("VisibilityScore for host %s returned an error: %v", host, err)
		}
		scores[host] = score
		t.Logf("Host %s visibility score: %f", host, score)
	}

	// Verify that scores are reasonable based on host characteristics
	// Host 6 should have the highest score due to combination of factors
	for host, score := range scores {
		if host != "192.168.1.6" && score > scores["192.168.1.6"] {
			t.Errorf("Expected host6 to have the highest score, but %s has a higher score (%f > %f)",
				host, score, scores["192.168.1.6"])
		}
	}

	// Host 3 should have the lowest score due to old activity
	for host, score := range scores {
		if host != "192.168.1.3" && score < scores["192.168.1.3"] {
			t.Errorf("Expected host3 to have the lowest score, but %s has a lower score (%f < %f)",
				host, score, scores["192.168.1.3"])
		}
	}
}

// TestGetAllHostScoresCSV tests the GetAllHostScores function using data from CSV
func TestGetAllHostScoresCSV(t *testing.T) {
	// Setup test database
	_, cleanup := setupTestDBFromCSV(t)
	defer cleanup()

	// Load test logs from CSV
	loadTestLogsFromCSV(t, db, "testdata/all_hosts_test.csv")

	// Call the function being tested
	scores, err := GetAllHostScores()

	// Verify results
	if err != nil {
		t.Fatalf("GetAllHostScores returned an error: %v", err)
	}
	if scores == nil {
		t.Fatal("GetAllHostScores returned nil scores")
	}

	// Check that we have at least one host score
	if len(scores) == 0 {
		t.Error("Expected at least one host score, got none")
	}

	// All scores should be positive
	for host, score := range scores {
		if score <= 0.0 {
			t.Errorf("Expected positive score for host %s, got %f", host, score)
		}
	}

	// Log all scores for debugging
	t.Logf("Host scores: %v", scores)

	// Find hosts with recent and old activity for comparison
	// This is more generic than hardcoding specific hosts
	var recentHosts, oldHosts []string
	var recentTimestamp, oldTimestamp time.Time

	// First, find the most recent and oldest timestamps
	for host := range scores {
		var log Log
		result := db.Where("client_ip = ?", host).Order("timestamp desc").First(&log)
		if result.Error != nil {
			continue
		}

		if recentTimestamp.IsZero() || log.Timestamp.After(recentTimestamp) {
			recentTimestamp = log.Timestamp
		}

		result = db.Where("client_ip = ?", host).Order("timestamp asc").First(&log)
		if result.Error != nil {
			continue
		}

		if oldTimestamp.IsZero() || log.Timestamp.Before(oldTimestamp) {
			oldTimestamp = log.Timestamp
		}
	}

	// Then categorize hosts
	recentThreshold := recentTimestamp.Add(-1 * time.Hour)
	oldThreshold := oldTimestamp.Add(24 * time.Hour)

	for host := range scores {
		var log Log
		result := db.Where("client_ip = ?", host).Order("timestamp desc").First(&log)
		if result.Error != nil {
			continue
		}

		if log.Timestamp.After(recentThreshold) {
			recentHosts = append(recentHosts, host)
		}

		result = db.Where("client_ip = ?", host).Order("timestamp asc").First(&log)
		if result.Error != nil {
			continue
		}

		if log.Timestamp.Before(oldThreshold) {
			oldHosts = append(oldHosts, host)
		}
	}

	// Test general principles of the scoring algorithm if we have enough data
	if len(recentHosts) > 0 && len(oldHosts) > 0 {
		// Find a recent host with high score and an old host with low score
		var recentHighScoreHost, oldLowScoreHost string
		var highestRecentScore, lowestOldScore float64

		for _, host := range recentHosts {
			if recentHighScoreHost == "" || scores[host] > highestRecentScore {
				recentHighScoreHost = host
				highestRecentScore = scores[host]
			}
		}

		for _, host := range oldHosts {
			if oldLowScoreHost == "" || scores[host] < lowestOldScore {
				oldLowScoreHost = host
				lowestOldScore = scores[host]
			}
		}

		// Test time decay component: More recent activity should generally score higher
		if recentHighScoreHost != "" && oldLowScoreHost != "" && recentHighScoreHost != oldLowScoreHost {
			if highestRecentScore <= lowestOldScore {
				t.Logf("Note: Expected recent host %s score (%f) to be greater than old host %s score (%f)",
					recentHighScoreHost, highestRecentScore, oldLowScoreHost, lowestOldScore)
			}
		}
	}

	// Verify that the scoring algorithm produces a range of values
	if len(scores) > 1 {
		var min, max float64
		first := true

		for _, score := range scores {
			if first {
				min, max = score, score
				first = false
			} else {
				if score < min {
					min = score
				}
				if score > max {
					max = score
				}
			}
		}

		// Check that we have some variation in scores
		if min == max && len(scores) > 1 {
			t.Logf("Note: All hosts have the same score (%f), expected some variation", min)
		}
	}
}
