package main

import (
	"fmt"
	"rss-reader/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		return
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		fmt.Printf("Failed to migrate: %v\n", err)
		return
	}

	fmt.Println("SQLite test passed!")
}
