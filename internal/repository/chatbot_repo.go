package repository

import (
	"github.com/yourusername/whatsapp-chatbot-go/internal/domain"
	"gorm.io/gorm"
)

type chatbotRepository struct {
	db *gorm.DB
}

func NewChatbotRepository(db *gorm.DB) ChatbotRepository {
	return &chatbotRepository{db: db}
}

func (r *chatbotRepository) FindByID(id string) (*domain.Chatbot, error) {
	var chatbot domain.Chatbot
	if err := r.db.Where("id = ?", id).First(&chatbot).Error; err != nil {
		return nil, err
	}
	return &chatbot, nil
}

func (r *chatbotRepository) FindByUserID(userID string) (*domain.Chatbot, error) {
	var chatbot domain.Chatbot
	if err := r.db.Where("user_id = ?", userID).First(&chatbot).Error; err != nil {
		return nil, err
	}
	return &chatbot, nil
}

func (r *chatbotRepository) Create(chatbot *domain.Chatbot) error {
	return r.db.Create(chatbot).Error
}

func (r *chatbotRepository) Update(chatbot *domain.Chatbot) error {
	return r.db.Save(chatbot).Error
}

func (r *chatbotRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Chatbot{}).Error
}

// ChatbotOptionRepository implementation
type chatbotOptionRepository struct {
	db *gorm.DB
}

func NewChatbotOptionRepository(db *gorm.DB) ChatbotOptionRepository {
	return &chatbotOptionRepository{db: db}
}

func (r *chatbotOptionRepository) FindByChatbotID(chatbotID string) ([]domain.ChatbotOption, error) {
	var options []domain.ChatbotOption
	if err := r.db.Where("chatbot_id = ?", chatbotID).Order("`order` ASC").Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *chatbotOptionRepository) FindByKey(chatbotID, optionKey string) (*domain.ChatbotOption, error) {
	var option domain.ChatbotOption
	if err := r.db.Where("chatbot_id = ? AND option_key = ?", chatbotID, optionKey).First(&option).Error; err != nil {
		return nil, err
	}
	return &option, nil
}

func (r *chatbotOptionRepository) Create(option *domain.ChatbotOption) error {
	return r.db.Create(option).Error
}

func (r *chatbotOptionRepository) Update(option *domain.ChatbotOption) error {
	return r.db.Save(option).Error
}

func (r *chatbotOptionRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.ChatbotOption{}).Error
}

// ConversationStateRepository implementation
type conversationStateRepository struct {
	db *gorm.DB
}

func NewConversationStateRepository(db *gorm.DB) ConversationStateRepository {
	return &conversationStateRepository{db: db}
}

func (r *conversationStateRepository) FindByUserAndChat(userID, chatID string) (*domain.ConversationState, error) {
	var state domain.ConversationState
	if err := r.db.Where("user_id = ? AND chat_id = ?", userID, chatID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *conversationStateRepository) Create(state *domain.ConversationState) error {
	return r.db.Create(state).Error
}

func (r *conversationStateRepository) Update(state *domain.ConversationState) error {
	return r.db.Save(state).Error
}