package database

import (
	"github.com/goSc4n/goSc4n/database/models"
)


// InitConfigSign used to init some default config
func InitConfigSign() {
	conObj := models.Configuration{
		Name:  "DefaultSign",
		Value: "*",
	}
	DB.Create(&conObj)
}


// UpdateDefaultSign update default sign
func UpdateDefaultSign(sign string) {
	var conObj models.Configuration
	DB.Where("name = ?", "DefaultSign").First(&conObj)
	conObj.Value = sign
	DB.Save(&conObj)
}

