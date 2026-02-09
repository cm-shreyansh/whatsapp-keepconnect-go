package domain

import "time"

type User struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Name            string     `json:"name" gorm:"type:varchar(255);not null"`
	Email           string     `json:"email" gorm:"type:varchar(255);uniqueIndex;"`
	EmailVerifiedAt *time.Time `json:"email_verified_at" gorm:"type:timestamp"`
	Password        string     `json:"-" gorm:"type:varchar(255);not null"`
	RememberToken   *string    `json:"remember_token" gorm:"type:varchar(100)"`
	ChatbotID       *string    `json:"chatbot_id" gorm:"type:varchar(255)"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}
