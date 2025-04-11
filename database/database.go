package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Temporarily using sqlite3
func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("./database/temp.sqlite3"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Unable to establish connection with the database: %s", err)
		return nil, err
	}

	db.AutoMigrate(&User{}, &Flat{})

	return db, nil
}
