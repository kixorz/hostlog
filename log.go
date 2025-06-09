package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net"
	"time"
)

type Log struct {
	gorm.Model
	ClientIP  string
	Hostname  string
	Content   string
	Priority  int
	Timestamp time.Time
}

type LogMap struct {
	gorm.Model
	ClientIP       string `gorm:"uniqueIndex"`
	HostnameField  string // Field name in logParts that maps to Log.Hostname
	ContentField   string // Field name in logParts that maps to Log.Content
	PriorityField  string // Field name in logParts that maps to Log.Priority
	TimestampField string // Field name in logParts that maps to Log.Timestamp
}

var db *gorm.DB

// InitDB initializes the database connection
func InitDB() (*gorm.DB, error) {
	var err error
	db, err = gorm.Open(sqlite.Open("logs.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	db.AutoMigrate(&Log{}, &LogMap{})

	return db, nil
}

// SaveLog saves a log entry to the database
func SaveLog(logParts map[string]interface{}) error {
	// Extract client IP:port from logParts
	clientString := getStringValue(logParts, "client")

	// Extract just the IP part for LogMap lookup
	clientIP := extractIP(clientString)

	// Create a new Log struct
	log := Log{
		ClientIP: clientIP,
	}

	// Try to find a LogMap configuration for this client IP (without port)
	logMaps, err := GetLogMapsByClientIP(clientIP)

	// If we found a mapping configuration, use it
	if err == nil && logMaps != nil {
		// Use the custom field mappings
		log.Hostname = getStringValue(logParts, logMaps.HostnameField)
		log.Content = getStringValue(logParts, logMaps.ContentField)
		log.Priority = getIntValue(logParts, logMaps.PriorityField)
		log.Timestamp = getTimeValue(logParts, logMaps.TimestampField)
	} else {
		// Fall back to default field mappings
		log.Hostname = getStringValue(logParts, "hostname")
		log.Content = getStringValue(logParts, "content")
		log.Priority = getIntValue(logParts, "priority")
		log.Timestamp = getTimeValue(logParts, "timestamp")
	}

	// Save to database
	return db.Create(&log).Error
}

func getStringValue(logParts map[string]interface{}, key string) string {
	if value, ok := logParts[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

func extractIP(clientString string) string {
	addr, err := net.ResolveUDPAddr("udp", clientString)
	if err != nil {
		return clientString
	}
	return addr.IP.String()
}

func getTimeValue(logParts map[string]interface{}, key string) time.Time {
	if value, ok := logParts[key]; ok {
		if timeValue, ok := value.(time.Time); ok {
			return timeValue
		}
	}
	return time.Time{}
}

func getIntValue(logParts map[string]interface{}, key string) int {
	if value, ok := logParts[key]; ok {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return 0
}

// GetLogMapsByClientIP retrieves the field mapping configuration for a specific client IP
func GetLogMapsByClientIP(clientIP string) (*LogMap, error) {
	var logMaps LogMap
	result := db.Where("client_ip = ?", clientIP).First(&logMaps)
	if result.Error != nil {
		return nil, result.Error
	}
	return &logMaps, nil
}

// CreateLogMaps creates a new field mapping configuration for a client IP
func CreateLogMaps(logMaps *LogMap) error {
	return db.Create(logMaps).Error
}

// UpdateLogMaps updates an existing field mapping configuration
func UpdateLogMaps(logMaps *LogMap) error {
	return db.Save(logMaps).Error
}

// DeleteLogMaps deletes a field mapping configuration for a client IP
func DeleteLogMaps(clientIP string) error {
	return db.Where("client_ip = ?", clientIP).Delete(&LogMap{}).Error
}

// GetFilteredLogs retrieves logs with filtering and pagination
func GetFilteredLogs(hosts []string, page int) ([]Log, int, error) {
	var logs []Log
	query := db.Order("created_at desc")

	if len(hosts) > 0 {
		query = query.Where("client_ip IN ?", hosts)
	}

	limit := 100
	offset := max(page, 0) * limit
	result := query.Offset(offset).Limit(limit).Find(&logs)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	var count int64
	result = query.Count(&count)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	maxPage := int((count+int64(limit)-1)/int64(limit) - 1)
	return logs, maxPage, result.Error
}

// GetUniqueClientIPs retrieves a list of unique client IPs from the database
func GetUniqueClientIPs() ([]string, error) {
	var clientIPs []string
	result := db.Model(&Log{}).Distinct("client_ip").Pluck("client_ip", &clientIPs)
	return clientIPs, result.Error
}

// CreateDefaultLogMaps creates a default LogMap entry for a client IP
// with standard field mappings
func CreateDefaultLogMaps(clientIP string) error {
	logMaps := LogMap{
		ClientIP:       clientIP,
		HostnameField:  "hostname",
		ContentField:   "content",
		PriorityField:  "priority",
		TimestampField: "timestamp",
	}
	return CreateLogMaps(&logMaps)
}

// Example of how to create a custom field mapping for a client IP
func ExampleCreateCustomLogMaps() {
	// Create a custom mapping for a specific client IP
	logMaps := LogMap{
		ClientIP:       "192.168.1.100",
		HostnameField:  "host",     // Custom field name for hostname
		ContentField:   "message",  // Custom field name for content
		PriorityField:  "severity", // Custom field name for priority
		TimestampField: "time",     // Custom field name for timestamp
	}

	// Save the custom mapping
	if err := CreateLogMaps(&logMaps); err != nil {
		fmt.Println("Error creating custom LogMap:", err)
	}
}
