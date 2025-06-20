package models

import (
	"gorm.io/gorm"
)

type LogTemplate struct {
	gorm.Model
	ClientIP       string `gorm:"uniqueIndex"`
	HostnameField  string // Field name in logParts that maps to Log.Hostname
	ContentField   string // Field name in logParts that maps to Log.Content
	PriorityField  string // Field name in logParts that maps to Log.Priority
	TimestampField string // Field name in logParts that maps to Log.Timestamp
}
