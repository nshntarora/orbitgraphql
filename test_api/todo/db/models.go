package db

import (
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID              uuid.UUID                `json:"id" gorm:"id"`
	Text            string                   `json:"text" gorm:"text"`
	Done            bool                     `json:"done" gorm:"done"`
	UserID          uuid.UUID                `json:"userId" gorm:"userId"`
	User            *User                    `json:"user" gorm:"-"`
	CreatedAt       time.Time                `json:"createdAt" gorm:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt" gorm:"updatedAt"`
	Meta            map[string]interface{}   `json:"meta" gorm:"-"`
	ActivityHistory []map[string]interface{} `json:"activityHistory" gorm:"-"`
}

type User struct {
	ID                      uuid.UUID              `json:"id" gorm:"id"`
	Name                    string                 `json:"name" gorm:"name"`
	Email                   string                 `json:"email" gorm:"email"`
	Username                string                 `json:"username" gorm:"username"`
	Tags                    []string               `json:"tags" gorm:"-"`
	CreatedAt               time.Time              `json:"createdAt" gorm:"createdAt"`
	UpdatedAt               time.Time              `json:"updatedAt" gorm:"updatedAt"`
	Todos                   []Todo                 `json:"todos" gorm:"-"`
	Meta                    map[string]interface{} `json:"meta" gorm:"-"`
	TodosCount              int                    `json:"todosCount" gorm:"-"`
	CompletionRate          float64                `json:"completionRate" gorm:"-"`
	CompletionRateLast7Days []float64              `json:"completionRateLast7Days" gorm:"-"`
	ActivityStreak7Days     []int                  `json:"activityStreak7Days" gorm:"-"`
}
