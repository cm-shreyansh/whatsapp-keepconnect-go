package handler

import (
	"fmt"

	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/middleware"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/service"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// SendTextMessage handles sending a text message
func (h *MessageHandler) SendTextMessage(c *fiber.Ctx) error {
	var req struct {
		UserID  string `json:"userId"`
		Phone   string `json:"phone"`
		Message string `json:"message"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if err := utils.ValidateRequired(map[string]string{
		"phone":   req.Phone,
		"message": req.Message,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Use userId from request if provided, otherwise from auth token
	userID := req.UserID
	if userID == "" {
		tokenUserID := middleware.GetUserID(c)
		userID = fmt.Sprintf("%d", tokenUserID)
	}

	resp, err := h.messageService.SendTextMessage(userID, req.Phone, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"message":    "Message sent successfully",
		"message_id": resp.MessageID,
		"timestamp":  resp.Timestamp,
	})
}

// SendMediaMessage handles sending a media message (image/video/document)
func (h *MessageHandler) SendMediaMessage(c *fiber.Ctx) error {
	var req struct {
		UserID   string `json:"userId"`
		Phone    string `json:"phone"`
		Caption  string `json:"caption"`
		MediaURL string `json:"mediaUrl"` // Changed from imageUrl to mediaUrl for clarity
		ImageURL string `json:"imageUrl"` // Keep for backwards compatibility
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Support both mediaUrl and imageUrl for backwards compatibility
	if req.MediaURL == "" {
		req.MediaURL = req.ImageURL
	}

	// Validate required fields
	if err := utils.ValidateRequired(map[string]string{
		"phone":    req.Phone,
		"mediaUrl": req.MediaURL,
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Use userId from request if provided, otherwise from auth token
	userID := req.UserID
	if userID == "" {
		tokenUserID := middleware.GetUserID(c)
		userID = fmt.Sprintf("%d", tokenUserID)
	}

	resp, err := h.messageService.SendMediaMessage(userID, req.Phone, req.MediaURL, req.Caption)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to send media",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"message":    "Media sent successfully",
		"message_id": resp.MessageID,
		"timestamp":  resp.Timestamp,
	})
}

// SendBulkTextMessages handles sending bulk text messages
func (h *MessageHandler) SendBulkTextMessages(c *fiber.Ctx) error {
	var req struct {
		UserID  string   `json:"userId"`
		Phones  []string `json:"phones"`
		Message string   `json:"message"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Phones) == 0 || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "phones and message are required",
		})
	}

	// Use userId from request if provided, otherwise from auth token
	userID := req.UserID
	if userID == "" {
		tokenUserID := middleware.GetUserID(c)
		userID = fmt.Sprintf("%d", tokenUserID)
	}

	results := h.messageService.SendBulkTextMessages(userID, req.Phones, req.Message)

	return c.JSON(fiber.Map{
		"success": true,
		"results": results,
	})
}

// SendBulkMediaMessages handles sending bulk media messages (image/video/document)
func (h *MessageHandler) SendBulkMediaMessages(c *fiber.Ctx) error {
	var req struct {
		UserID   string   `json:"userId"`
		Phones   []string `json:"phones"`
		Message  string   `json:"message"`
		MediaURL string   `json:"mediaUrl"` // New field name
		ImageURL string   `json:"imageUrl"` // Keep for backwards compatibility
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Support both mediaUrl and imageUrl for backwards compatibility
	if req.MediaURL == "" {
		req.MediaURL = req.ImageURL
	}

	if len(req.Phones) == 0 || req.Message == "" || req.MediaURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "phones, message, and mediaUrl (or imageUrl) are required",
		})
	}

	// Use userId from request if provided, otherwise from auth token
	userID := req.UserID
	if userID == "" {
		tokenUserID := middleware.GetUserID(c)
		userID = fmt.Sprintf("%d", tokenUserID)
	}

	results := h.messageService.SendBulkMediaMessages(userID, req.Phones, req.MediaURL, req.Message)

	return c.JSON(fiber.Map{
		"success": true,
		"results": results,
	})
}
