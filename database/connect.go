package database

import (
	"github.com/goSc4n/goSc4n/database/models"
	"github.com/jinzhu/gorm"
	// load driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// DB global DB variable
var DB *gorm.DB

// InitDB init DB connection
func InitDB(DBPath string) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", DBPath)
	db.LogMode(false)

	if err == nil {
		DB = db
		db.AutoMigrate(&models.Scans{})
		db.AutoMigrate(&models.Record{})
		db.AutoMigrate(&models.Signature{})
		db.AutoMigrate(&models.User{})
		db.AutoMigrate(&models.Configuration{})
		return db, err
	}
	return nil, err
}
