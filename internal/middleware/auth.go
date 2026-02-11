package middleware

import (
	"strings"
	"time"

	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/config"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	// UserID    uint   `json:"user_id"`
	ChatbotID string `json:"chatbot_id"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	jwtSecret string
	userRepo  repository.UserRepository
}

func NewAuthMiddleware(cfg *config.Config, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: cfg.JWT.Secret,
		userRepo:  userRepo,
	}
}

// Auth validates JWT token and adds user info to context
func (am *AuthMiddleware) Auth(c *fiber.Ctx) error {
	// Get authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid authorization format. Use: Bearer <token>",
		})
	}

	tokenString := parts[1]

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(am.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	user, error := am.userRepo.FindByID(claims.UserID)
	if error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Add user info to context
	userId, err := claims.GetSubject()
	c.Locals("user_id", userId)
	c.Locals("chatbot_id", user.ChatbotID)

	return c.Next()
}

// GenerateToken generates a JWT token for a user
func (am *AuthMiddleware) GenerateToken(userID uint, chatbotID string) (string, error) {
	claims := Claims{
		UserID:    userID,
		ChatbotID: chatbotID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 30)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.jwtSecret))
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) uint {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return 0
	}
	return userID
}

// GetChatbotID extracts chatbot ID from context
func GetChatbotID(c *fiber.Ctx) string {
	chatbotID, ok := c.Locals("chatbot_id").(string)
	if !ok {
		return ""
	}
	return chatbotID
}
