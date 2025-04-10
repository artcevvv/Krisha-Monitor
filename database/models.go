package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ChatID      int64
	City        string
	Region      string
	PricingFrom int
	PricingTo   int
}

type Flat struct {
	gorm.Model
	user      User
	url       string
	city      string
	region    string
	priceFrom int
	priceTo   int
}
