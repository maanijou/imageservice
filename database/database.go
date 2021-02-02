package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var GlobalDB *gorm.DB

// InitDatabase creates a sqlite db
func InitDatabase(dbName string) (err error) {
	GlobalDB, err = gorm.Open(sqlite.Open(fmt.Sprintf("%s.db", dbName)), &gorm.Config{})
	if err != nil {
		return
	}
	return
}
