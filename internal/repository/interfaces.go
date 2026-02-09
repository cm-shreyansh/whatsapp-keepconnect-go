package repository

import "github.com/yourusername/whatsapp-chatbot-go/internal/domain"

// UserRepository defines the interface for user data operations
type UserRepository interface {
	FindByID(id uint) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
}

// ChatbotRepository defines the interface for chatbot data operations
type ChatbotRepository interface {
	FindByID(id string) (*domain.Chatbot, error)
	FindByUserID(userID string) (*domain.Chatbot, error)
	Create(chatbot *domain.Chatbot) error
	Update(chatbot *domain.Chatbot) error
	Delete(id string) error
}

// ChatbotOptionRepository defines the interface for chatbot option data operations
type ChatbotOptionRepository interface {
	FindByChatbotID(chatbotID string) ([]domain.ChatbotOption, error)
	FindByKey(chatbotID, optionKey string) (*domain.ChatbotOption, error)
	Create(option *domain.ChatbotOption) error
	Update(option *domain.ChatbotOption) error
	Delete(id string) error
}

// ConversationStateRepository defines the interface for conversation state data operations
type ConversationStateRepository interface {
	FindByUserAndChat(userID, chatID string) (*domain.ConversationState, error)
	Create(state *domain.ConversationState) error
	Update(state *domain.ConversationState) error
}