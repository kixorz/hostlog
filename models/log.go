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

func GetAllHosts() ([]string, error) {
	var hosts []string
	result := DB.Model(&Log{}).Distinct("client_ip").Pluck("client_ip", &hosts)
	return hosts, result.Error
}

func GetLogs(host string, timeWindow time.Time) ([]Log, error) {
	var logs []Log
	result := DB.Model(&Log{}).Where("client_ip = ? AND timestamp > ?", host, timeWindow).Find(&logs)
	if result.Error != nil {
		return logs, result.Error
	}
	return logs, nil
}

func GetCount(host string, oneHourAgo time.Time) (int64, error) {
	var count int64
	result := DB.Model(&Log{}).Where("client_ip = ? AND timestamp > ?", host, oneHourAgo).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func GetFirst(host string) (Log, error) {
	var log Log
	result := DB.Where("client_ip = ?", host).Order("created_at desc").First(&log)
	if result.Error != nil {
		return log, result.Error
	}
	return log, nil
}
