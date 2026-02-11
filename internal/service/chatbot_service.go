package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/domain"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/repository"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/utils"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/pkg/whatsmeow_client"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type ChatbotService struct {
	chatbotRepo      repository.ChatbotRepository
	optionRepo       repository.ChatbotOptionRepository
	conversationRepo repository.ConversationStateRepository
	userRepo         repository.UserRepository
	waManager        *whatsmeow_client.Manager
}

func NewChatbotService(
	chatbotRepo repository.ChatbotRepository,
	optionRepo repository.ChatbotOptionRepository,
	conversationRepo repository.ConversationStateRepository,
	userRepo repository.UserRepository,
	waManager *whatsmeow_client.Manager,
) *ChatbotService {
	return &ChatbotService{
		chatbotRepo:      chatbotRepo,
		optionRepo:       optionRepo,
		conversationRepo: conversationRepo,
		userRepo:         userRepo,
		waManager:        waManager,
	}
}

// HandleMessage implements the whatsmeow_client.EventHandler interface
func (s *ChatbotService) HandleMessage(userID string, message interface{}) {
	s.HandleIncomingMessage(userID, message)
}

// HandleIncomingMessage processes incoming WhatsApp messages for chatbot
func (s *ChatbotService) HandleIncomingMessage(userID string, messageEvt interface{}) {
	// Extract message event
	log.Println("\n~ Incoming message")
	fmt.Print("\n Yeaaa \n")
	msgEvent, err := whatsmeow_client.ExtractMessageEvent(messageEvt)
	if err != nil {
		return // Not a message event
	}

	// Ignore messages sent by the bot itself
	if msgEvent.FromMe {
		return
	}

	chatID := msgEvent.From
	messageBody := strings.TrimSpace(msgEvent.Body)
	log.Println("\n~ " + messageBody + "\n")
	fmt.Print("\n" + msgEvent.Body + "\n")
	fmt.Print("\n" + msgEvent.From + "\n")

	if messageBody == "" {
		return
	}

	// Get chatbot for this user
	chatbot, err := s.chatbotRepo.FindByUserID(userID)
	if err != nil {
		log.Printf("No active chatbot for user %s: %v", userID, err)
		return
	}

	if !chatbot.IsActive {
		log.Printf("Chatbot is inactive for user %s", userID)
		return
	}

	// Get WhatsApp client
	clientData, exists := s.waManager.GetClient(userID)
	if !exists {
		log.Printf("WhatsApp client not found for user %s", userID)
		return
	}

	// fmt.Print("\n Greetings boii \n")
	// Check if it's a greeting
	utils.IsGreeting(messageBody)
	fmt.Print("THIS IS IT BREOOOOOO\n")
	if utils.IsGreeting(messageBody) {
		log.Printf("It is Greetings %s", userID)
		s.handleGreeting(chatbot, chatID, clientData)
		s.updateConversationState(userID, chatID)
		return
	}

	// Check if message matches any option key
	options, err := s.optionRepo.FindByChatbotID(chatbot.ID)
	if err != nil {
		log.Printf("Failed to fetch options: %v", err)
		return
	}

	// Find matching option
	var matchedOption *domain.ChatbotOption
	for i := range options {
		if strings.EqualFold(options[i].OptionKey, messageBody) {
			matchedOption = &options[i]
			break
		}
	}

	if matchedOption != nil {
		s.handleOptionResponse(matchedOption, chatID, clientData)
		s.updateConversationState(userID, chatID)
	}

	// If no match, don't reply (as per requirement)
}

func (s *ChatbotService) handleGreeting(chatbot *domain.Chatbot, chatID string, clientData *whatsmeow_client.ClientData) {
	jid, err := types.ParseJID(chatID)
	if err != nil {
		log.Printf("Failed to parse JID: %v", err)
		return
	}

	// Send media if available, otherwise send text
	if chatbot.MediaURL != nil && *chatbot.MediaURL != "" {
		s.sendMediaMessage(clientData, jid, *chatbot.MediaURL, chatbot.WelcomeMessage)
	} else {
		s.sendTextMessage(clientData, jid, chatbot.WelcomeMessage)
	}
}

func (s *ChatbotService) handleOptionResponse(option *domain.ChatbotOption, chatID string, clientData *whatsmeow_client.ClientData) {
	jid, err := types.ParseJID(chatID)
	if err != nil {
		log.Printf("Failed to parse JID: %v", err)
		return
	}

	// Send media with caption if available, otherwise send text
	if option.MediaURL != nil && *option.MediaURL != "" {
		s.sendMediaMessage(clientData, jid, *option.MediaURL, option.Answer)
	} else {
		s.sendTextMessage(clientData, jid, option.Answer)
	}
}

func (s *ChatbotService) sendTextMessage(clientData *whatsmeow_client.ClientData, jid types.JID, message string) {
	_, err := clientData.Client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: proto.String(message),
	})
	if err != nil {
		log.Printf("Failed to send text message: %v", err)
	}
}

func (s *ChatbotService) sendMediaMessage(clientData *whatsmeow_client.ClientData, jid types.JID, mediaURL, caption string) {
	// Download media
	resp, err := http.Get(mediaURL)
	if err != nil {
		log.Printf("Failed to download media: %v", err)
		s.sendTextMessage(clientData, jid, caption) // Fallback to text
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read media: %v", err)
		s.sendTextMessage(clientData, jid, caption) // Fallback to text
		return
	}

	// Upload media to WhatsApp
	uploaded, err := clientData.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		log.Printf("Failed to upload media: %v", err)
		s.sendTextMessage(clientData, jid, caption) // Fallback to text
		return
	}

	// Send image message
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       proto.String(caption),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}

	_, err = clientData.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		log.Printf("Failed to send media message: %v", err)
	}
}

func (s *ChatbotService) updateConversationState(userID, chatID string) {
	existing, err := s.conversationRepo.FindByUserAndChat(userID, chatID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new conversation state
			newState := &domain.ConversationState{
				ID:              utils.GenerateID("conv_"),
				UserID:          userID,
				ChatID:          chatID,
				LastMessageTime: time.Now(),
			}
			s.conversationRepo.Create(newState)
		}
	} else {
		// Update existing state
		existing.LastMessageTime = time.Now()
		s.conversationRepo.Update(existing)
	}
}

// SetChatbotInactive deactivates a chatbot
func (s *ChatbotService) SetChatbotInactive(userID string) error {
	chatbot, err := s.chatbotRepo.FindByUserID(userID)
	if err != nil {
		return fmt.Errorf("chatbot not found: %w", err)
	}

	chatbot.IsActive = false
	chatbot.UpdatedAt = time.Now()

	return s.chatbotRepo.Update(chatbot)
}
