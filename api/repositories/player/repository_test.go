package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewPlayerRepository(t *testing.T) {
	repository := NewPlayerRepository(&gorm.DB{})
	assert.NotNil(t, repository)
}
