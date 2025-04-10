package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Temporarily using sqlite3
func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("temp.db"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Unable to establish connection with the database: %s", err)
		return nil, err
	}

	return db, nil
}

func MigrateAll() error {
	db, err := InitDB()
	if err != nil {
		fmt.Printf("Unable to establish connection with the database: %s", err)
		return err
	}

	err = db.AutoMigrate(&User{}, &Flat{})
	if err != nil {
		fmt.Printf("Unable to migrate models: %s", err)
		return err
	}

	return nil
}
