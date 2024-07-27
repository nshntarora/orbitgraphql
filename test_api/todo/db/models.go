package db

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID        uuid.UUID `json:"id" gorm:"id"`
	Text      string    `json:"text" gorm:"text"`
	Done      bool      `json:"done" gorm:"done"`
	UserID    uuid.UUID `json:"userId" gorm:"userId"`
	User      *User     `json:"user" gorm:"user"`
	CreatedAt time.Time `json:"createdAt" gorm:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"updatedAt"`
}

type User struct {
	ID        uuid.UUID `json:"id" gorm:"id"`
	Name      string    `json:"name" gorm:"name"`
	Email     string    `json:"email" gorm:"email"`
	Username  string    `json:"username" gorm:"username"`
	CreatedAt time.Time `json:"createdAt" gorm:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"updatedAt"`
}
