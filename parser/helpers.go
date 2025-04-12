package parser

import (
	"database"
	"fmt"

	"gorm.io/gorm"
)

func FormURL(db *gorm.DB, chatID int64) string {
	var user database.User
	err := db.First(&user, "chat_id = ?", chatID).Error
	if err != nil {
		fmt.Printf("Error fetching user data: %v\n", err)
		return ""
	}

	if user.Region != "" {
		return fmt.Sprintf("%s/%s/?das[price][from]=%d&das[price][to]=%d", KRISHA_ARENDA_URL, user.Region, user.PricingFrom, user.PricingTo)
	}

	return fmt.Sprintf("%s/%s/", KRISHA_ARENDA_URL, user.City)
}
