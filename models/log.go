package models

import (
	"gorm.io/gorm"
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
