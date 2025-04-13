package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ChatID      int64 `gorm:"primaryKey"`
	City        string
	Region      string
	PricingFrom int
	PricingTo   int
}

type Flat struct {
	gorm.Model
	UserID      uint `gorm:"index"`
	User        User
	Title       string
	Price       string
	Location    string
	Description string
	Link        string
	ImageURL    string
	Date        string
	Area        string
	Floor       string
	Rooms       string
	Page        int
}
