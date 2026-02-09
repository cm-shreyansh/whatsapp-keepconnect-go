package domain

import "time"

type Chatbot struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID         string    `json:"user_id" gorm:"type:varchar(255);uniqueIndex;not null"`
	WelcomeMessage string    `json:"welcome_message" gorm:"type:text;not null"`
	MediaURL       *string   `json:"media_url" gorm:"type:varchar(255)"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Chatbot) TableName() string {
	return "chatbots"
}

type ChatbotOption struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	ChatbotID   string    `json:"chatbot_id" gorm:"type:varchar(255);not null;index"`
	OptionKey   string    `json:"option_key" gorm:"type:varchar(50);not null"`
	OptionLabel string    `json:"option_label" gorm:"type:text;not null"`
	Answer      string    `json:"answer" gorm:"type:text;not null"`
	MediaURL    *string   `json:"media_url" gorm:"type:text"`
	MediaType   *string   `json:"media_type" gorm:"type:varchar(50)"`
	Order       int       `json:"order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ChatbotOption) TableName() string {
	return "chatbot_options"
}

type ConversationState struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID          string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	ChatID          string    `json:"chat_id" gorm:"type:varchar(255);not null;index"`
	LastMessageTime time.Time `json:"last_message_time" gorm:"autoCreateTime"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ConversationState) TableName() string {
	return "conversation_states"
}