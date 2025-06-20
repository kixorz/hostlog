package models

import (
	"gorm.io/gorm"
)

type LogField struct {
	gorm.Model
	ClientIP  string `gorm:"index"`
	FieldName string `gorm:"index"`
	Count     int
}

func SaveLogFields(clientIP string, logParts map[string]interface{}) {
	for fieldName := range logParts {
		var logField LogField
		result := db.Where("client_ip = ? AND field_name = ?", clientIP, fieldName).First(&logField)
		if result.Error != nil {
			logField = LogField{
				ClientIP:  clientIP,
				FieldName: fieldName,
				Count:     1,
			}
			db.Create(&logField)
		} else {
			logField.Count++
			db.Save(&logField)
		}
	}
}

func GetLogFieldsByClientIP(clientIP string) ([]LogField, error) {
	var logFields []LogField
	result := db.Where("client_ip = ?", clientIP).Order("count desc").Find(&logFields)
	if result.Error != nil {
		return nil, result.Error
	}
	return logFields, nil
}
