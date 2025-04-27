package repositories

import (
	"fmt"
	"goleague/pkg/database"

	"gorm.io/gorm"
)

// Public Interface.
type TierlistRepository interface{}

// Tierlist repository structure.
type tierlistRepository struct {
	db *gorm.DB
}

// Create a tierlist repository.
func NewTierlistRepository() (TierlistRepository, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("couldn't get database connection: %w", err)
	}
	return &tierlistRepository{db: db}, nil
}
