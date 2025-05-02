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

func SaveData(db *gorm.DB, data User) error {
	if err := db.Save(&data).Error; err != nil {
		return err
	}

	return nil
}

func GetUser(db *gorm.DB, chatID int64) (*User, error) {
	var user User
	if err := db.Where("chat_id = ?", chatID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func SaveFlats(db *gorm.DB, flats []Flat) error {
	if err := db.Save(&flats).Error; err != nil {
		return fmt.Errorf("error saving flats: %v", err)
	}
	return nil
}

func GetFlatsByUser(db *gorm.DB, userID int64) ([]Flat, error) {
	var flats []Flat
	if err := db.Where("user_id = ?", userID).Find(&flats).Error; err != nil {
		return nil, err
	}
	return flats, nil
}

func GetFlatsByPage(db *gorm.DB, page int) ([]Flat, error) {
	var flats []Flat
	if err := db.Where("page = ?", page).Find(&flats).Error; err != nil {
		return nil, err
	}
	return flats, nil
}

func DeleteOldFlats(db *gorm.DB, flatsToDelete []uint) error {
	if len(flatsToDelete) == 0 {
		return nil
	}
	if err := db.Delete(&Flat{}, flatsToDelete).Error; err != nil {
		return fmt.Errorf("[deleteOldFlats] Error deleting old flats: %v", err)
	}
	fmt.Printf("[deleteOldFlats] Deleted %d outdated flats\n", len(flatsToDelete))
	return nil
}

func GetData(db *gorm.DB, chatID string) (interface{}, bool) {
	data, status := db.Get(chatID)

	return data, status
}

func UpdateData(db *gorm.DB, userID int64, newRegion string, priceFrom, priceTo int) error {
	user, err := GetUser(db, userID)
	if err != nil {
		return err
	}

	if err := db.Model(user).Updates(map[string]interface{}{
		"region":       newRegion,
		"pricing_from": priceFrom,
		"pricing_to":   priceTo,
	}).Error; err != nil {
		return fmt.Errorf("error updating user data: %v", err)
	}

	return nil
}