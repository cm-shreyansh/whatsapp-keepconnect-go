package handler

import (
	// "fmt"
	"time"

	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/domain"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/middleware"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/repository"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ChatbotHandler struct {
	chatbotRepo repository.ChatbotRepository
	optionRepo  repository.ChatbotOptionRepository
	userRepo    repository.UserRepository
	db          *gorm.DB
}

func NewChatbotHandler(
	chatbotRepo repository.ChatbotRepository,
	optionRepo repository.ChatbotOptionRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
) *ChatbotHandler {
	return &ChatbotHandler{
		chatbotRepo: chatbotRepo,
		optionRepo:  optionRepo,
		userRepo:    userRepo,
		db:          db,
	}
}

// CreateOrUpdateChatbot creates or updates a chatbot
func (h *ChatbotHandler) CreateOrUpdateChatbot(c *fiber.Ctx) error {
	var req struct {
		UserID         string  `json:"userId"`
		WelcomeMessage string  `json:"welcomeMessage"`
		IsActive       *bool   `json:"isActive"`
		MediaURL       *string `json:"mediaUrl"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	chatbotID := middleware.GetChatbotID(c)
	accountUserID := middleware.GetUserID(c)

	// Validate required fields
	if err := utils.ValidateRequired(map[string]string{
		"userId":         req.UserID,
		"welcomeMessage": req.WelcomeMessage,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var chatbot *domain.Chatbot
	var isUpdate bool

	// Check if chatbot exists
	if chatbotID != "" {
		existing, err := h.chatbotRepo.FindByID(chatbotID)
		if err == nil {
			// Update existing
			existing.WelcomeMessage = req.WelcomeMessage
			existing.UserID = req.UserID
			existing.MediaURL = req.MediaURL
			if req.IsActive != nil {
				existing.IsActive = *req.IsActive
			} else {
				existing.IsActive = true
			}
			existing.UpdatedAt = time.Now()

			if err := h.chatbotRepo.Update(existing); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Failed to update chatbot",
					"details": err.Error(),
				})
			}

			chatbot = existing
			isUpdate = true
		}
	}

	// Create new chatbot if not updating
	if chatbot == nil {
		newChatbotID := utils.GenerateID("bot_")

		// Start transaction
		tx := h.db.Begin()

		newChatbot := &domain.Chatbot{
			ID:             newChatbotID,
			UserID:         req.UserID,
			WelcomeMessage: req.WelcomeMessage,
			MediaURL:       req.MediaURL,
			IsActive:       true,
		}
		if req.IsActive != nil {
			newChatbot.IsActive = *req.IsActive
		}

		if err := tx.Create(newChatbot).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create chatbot",
				"details": err.Error(),
			})
		}

		// Update user's chatbot_id
		if err := tx.Model(&domain.User{}).Where("id = ?", accountUserID).Update("chatbot_id", newChatbotID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to link chatbot to user",
				"details": err.Error(),
			})
		}

		tx.Commit()
		chatbot = newChatbot
	}

	message := "Chatbot created"
	if isUpdate {
		message = "Chatbot updated"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"chatbot": chatbot,
		"message": message,
	})
}

// GetChatbot retrieves a chatbot with its options
func (h *ChatbotHandler) GetChatbot(c *fiber.Ctx) error {
	chatbotID := middleware.GetChatbotID(c)

	if chatbotID == "" {
		return c.JSON(fiber.Map{
			"chatbot": nil,
			"options": nil,
		})
	}

	chatbot, err := h.chatbotRepo.FindByID(chatbotID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Chatbot not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch chatbot",
			"details": err.Error(),
		})
	}

	options, err := h.optionRepo.FindByChatbotID(chatbot.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch options",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"chatbot": chatbot,
		"options": options,
	})
}

// CreateOrUpdateOption creates or updates a chatbot option
func (h *ChatbotHandler) CreateOrUpdateOption(c *fiber.Ctx) error {
	var req struct {
		OptionKey   string  `json:"optionKey"`
		OptionLabel string  `json:"optionLabel"`
		Answer      string  `json:"answer"`
		MediaURL    *string `json:"mediaUrl"`
		MediaType   *string `json:"mediaType"`
		Order       *int    `json:"order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	chatbotID := middleware.GetChatbotID(c)

	// Validate required fields
	if err := utils.ValidateRequired(map[string]string{
		"optionKey":   req.OptionKey,
		"optionLabel": req.OptionLabel,
		"answer":      req.Answer,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if chatbot exists
	if chatbotID == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found. Create a chatbot first.",
		})
	}

	_, err := h.chatbotRepo.FindByID(chatbotID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found. Create a chatbot first.",
		})
	}

	// Check if option exists
	existing, err := h.optionRepo.FindByKey(chatbotID, req.OptionKey)
	var option *domain.ChatbotOption

	if err == nil {
		// Update existing option
		existing.OptionLabel = req.OptionLabel
		existing.Answer = req.Answer
		existing.MediaURL = req.MediaURL
		existing.MediaType = req.MediaType
		if req.Order != nil {
			existing.Order = *req.Order
		}
		existing.UpdatedAt = time.Now()

		if err := h.optionRepo.Update(existing); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to update option",
				"details": err.Error(),
			})
		}

		option = existing
	} else {
		// Create new option
		order := 0
		if req.Order != nil {
			order = *req.Order
		}

		newOption := &domain.ChatbotOption{
			ID:          utils.GenerateID("opt_"),
			ChatbotID:   chatbotID,
			OptionKey:   req.OptionKey,
			OptionLabel: req.OptionLabel,
			Answer:      req.Answer,
			MediaURL:    req.MediaURL,
			MediaType:   req.MediaType,
			Order:       order,
		}

		if err := h.optionRepo.Create(newOption); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create option",
				"details": err.Error(),
			})
		}

		option = newOption
	}

	message := "Option created"
	if existing != nil {
		message = "Option updated"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"option":  option,
		"message": message,
	})
}

// DeleteOption deletes a chatbot option
func (h *ChatbotHandler) DeleteOption(c *fiber.Ctx) error {
	userID := c.Params("userId")
	optionKey := c.Params("optionKey")

	if userID == "" || optionKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId and optionKey are required",
		})
	}

	chatbot, err := h.chatbotRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found",
		})
	}

	option, err := h.optionRepo.FindByKey(chatbot.ID, optionKey)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Option not found",
		})
	}

	if err := h.optionRepo.Delete(option.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to delete option",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Option deleted",
	})
}

// ToggleChatbot toggles chatbot active status
func (h *ChatbotHandler) ToggleChatbot(c *fiber.Ctx) error {
	userID := c.Params("userId")

	var req struct {
		IsActive bool `json:"isActive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	chatbot, err := h.chatbotRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found",
		})
	}

	chatbot.IsActive = req.IsActive
	chatbot.UpdatedAt = time.Now()

	if err := h.chatbotRepo.Update(chatbot); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update chatbot",
			"details": err.Error(),
		})
	}

	message := "Chatbot activated"
	if !req.IsActive {
		message = "Chatbot deactivated"
	}

	return c.JSON(fiber.Map{
		"success": true,
		"chatbot": chatbot,
		"message": message,
	})
}

// DeleteChatbot deletes a chatbot
func (h *ChatbotHandler) DeleteChatbot(c *fiber.Ctx) error {
	userID := c.Params("userId")

	chatbot, err := h.chatbotRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found",
		})
	}

	if err := h.chatbotRepo.Delete(chatbot.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to delete chatbot",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatbot deleted",
	})
}
