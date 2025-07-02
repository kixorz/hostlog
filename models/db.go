package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net"
	"os"
	"time"
)

var db *gorm.DB
var DBPath string

// DefaultDBPath is the default path for the SQLite database file
const DefaultDBPath = "logs.db"

func InitDB() (*gorm.DB, error) {
	// Check if DB path is provided via environment variable
	DBPath = os.Getenv("AKLOG_DB_PATH")
	if DBPath == "" {
		DBPath = DefaultDBPath
	}

	var err error
	db, err = gorm.Open(sqlite.Open(DBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Log{}, &LogTemplate{}, &LogField{})

	return db, nil
}

func SaveLog(logParts map[string]interface{}) error {
	clientString := GetStringValue(logParts, "client")
	clientIP := ExtractIP(clientString)
	log := Log{
		ClientIP: clientIP,
	}

	// Try to find a LogTemplate configuration for this client IP (without port)
	logMaps, err := GetLogMapsByClientIP(clientIP)

	// If we found a mapping configuration, use it
	if err == nil && logMaps != nil {
		// Use the custom field mappings
		log.Hostname = GetStringValue(logParts, logMaps.HostnameField)
		log.Content = GetStringValue(logParts, logMaps.ContentField)
		log.Priority = GetIntValue(logParts, logMaps.PriorityField)
		log.Timestamp = GetTimeValue(logParts, logMaps.TimestampField)
	} else {
		// Fall back to default field mappings
		log.Hostname = GetStringValue(logParts, "hostname")
		log.Content = GetStringValue(logParts, "content")
		log.Priority = GetIntValue(logParts, "priority")
		log.Timestamp = GetTimeValue(logParts, "timestamp")
	}

	go SaveLogFields(clientIP, logParts)

	return db.Create(&log).Error
}

func GetStringValue(logParts map[string]interface{}, key string) string {
	if value, ok := logParts[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

func ExtractIP(clientString string) string {
	addr, err := net.ResolveUDPAddr("udp", clientString)
	if err != nil {
		return clientString
	}
	return addr.IP.String()
}

func GetTimeValue(logParts map[string]interface{}, key string) time.Time {
	if value, ok := logParts[key]; ok {
		if timeValue, ok := value.(time.Time); ok {
			return timeValue
		}
	}
	return time.Time{}
}

func GetIntValue(logParts map[string]interface{}, key string) int {
	if value, ok := logParts[key]; ok {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return 0
}

func GetLogMapsByClientIP(clientIP string) (*LogTemplate, error) {
	var logMaps LogTemplate
	result := db.Where("client_ip = ?", clientIP).First(&logMaps)
	if result.Error != nil {
		return nil, result.Error
	}
	return &logMaps, nil
}

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
