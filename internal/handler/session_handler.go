package handler

import (
	"fmt"

	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/middleware"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/service"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/pkg/whatsmeow_client"
	"github.com/gofiber/fiber/v2"
)

type SessionHandler struct {
	waManager      *whatsmeow_client.Manager
	chatbotService *service.ChatbotService
}

func NewSessionHandler(waManager *whatsmeow_client.Manager, chatbotService *service.ChatbotService) *SessionHandler {
	return &SessionHandler{
		waManager:      waManager,
		chatbotService: chatbotService,
	}
}

// InitSession initializes a new WhatsApp session
func (h *SessionHandler) InitSession(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in token",
		})
	}

	userIDStr := fmt.Sprintf("%d", userID) // Convert uint to string
	clientData, err := h.waManager.InitializeClient(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to initialize session",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session initialized. Scan QR code to authenticate.",
		"status":  clientData.GetStatus(),
	})
}

// GetQRCode returns the QR code for authentication
func (h *SessionHandler) GetQRCode(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId is required",
		})
	}

	clientData, exists := h.waManager.GetClient(userID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Session not found. Initialize session first.",
		})
	}

	qrCode := clientData.GetQRCode()
	status := clientData.GetStatus()

	if qrCode == "" || qrCode == "undefined" {
		return c.JSON(fiber.Map{
			"status":  status,
			"message": "QR code not available yet or already authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"status": status,
		"qr":     qrCode,
	})
}

// GetStatus returns the current session status
func (h *SessionHandler) GetStatus(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId is required",
		})
	}

	clientData, exists := h.waManager.GetClient(userID)
	if !exists {
		return c.JSON(fiber.Map{
			"status":       whatsmeow_client.StatusNotInitialized,
			"is_logged_in": false,
			"message":      "Session not found",
		})
	}

	status := clientData.GetStatus()

	return c.JSON(fiber.Map{
		"status":       status,
		"is_logged_in": status == whatsmeow_client.StatusReady,
	})
}

// Logout logs out and destroys the session
func (h *SessionHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId is required",
		})
	}

	// Set chatbot to inactive
	if err := h.chatbotService.SetChatbotInactive(req.UserID); err != nil {
		// Log but don't fail the logout
		c.Append("X-Warning", "Failed to deactivate chatbot")
	}

	// Logout from WhatsApp
	if err := h.waManager.LogoutClient(req.UserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to logout",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetAllSessions returns all active sessions
func (h *SessionHandler) GetAllSessions(c *fiber.Ctx) error {
	sessions := h.waManager.GetAllSessions()

	activeSessions := 0
	for _, session := range sessions {
		if isLoggedIn, ok := session["is_logged_in"].(bool); ok && isLoggedIn {
			activeSessions++
		}
	}

	return c.JSON(fiber.Map{
		"sessions":        sessions,
		"total_sessions":  len(sessions),
		"active_sessions": activeSessions,
	})
}
