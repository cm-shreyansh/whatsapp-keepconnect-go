package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/config"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/database"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/handler"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/middleware"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/repository"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/internal/service"
	"github.com/cm-shreyansh/whatsapp-keepconnect-go/pkg/whatsmeow_client"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// // Run migrations
	// if err := database.AutoMigrate(db); err != nil {
	// 	log.Fatalf("Failed to migrate database: %v", err)
	// }

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	chatbotRepo := repository.NewChatbotRepository(db)
	optionRepo := repository.NewChatbotOptionRepository(db)
	conversationRepo := repository.NewConversationStateRepository(db)

	// Initialize services (temporary placeholders for WhatsApp manager)
	chatbotService := service.NewChatbotService(chatbotRepo, optionRepo, conversationRepo, userRepo, nil)

	// Initialize WhatsApp manager with chatbot service as event handler
	waManager, err := whatsmeow_client.NewManager(cfg.WhatsApp.DBPath, chatbotService)
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp manager: %v", err)
	}

	// Start periodic metadata saving (every 5 minutes)
	waManager.StartMetadataSaver(5 * time.Minute)

	// Update chatbot service with the WhatsApp manager
	chatbotService = service.NewChatbotService(chatbotRepo, optionRepo, conversationRepo, userRepo, waManager)
	messageService := service.NewMessageService(waManager)

	// Initialize handlers
	sessionHandler := handler.NewSessionHandler(waManager, chatbotService)
	messageHandler := handler.NewMessageHandler(messageService)
	chatbotHandler := handler.NewChatbotHandler(chatbotRepo, optionRepo, userRepo, db)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Health check
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now(),
		})
	})

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "WhatsApp Multi-Account API - Go Edition",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /api/session/init",
				"GET /api/session/qr/:userId",
				"GET /api/session/status/:userId",
				"POST /api/session/logout",
				"POST /api/message/send",
				"POST /api/message/send-many",
				"POST /api/message/send-media",
				"POST /api/message/send-many-image",
				"GET /api/sessions",
				"--- CHATBOT ENDPOINTS ---",
				"POST /api/chatbot",
				"GET /api/chatbot",
				"POST /api/chatbot/option",
				"DELETE /api/chatbot/option/:userId/:optionKey",
				"PATCH /api/chatbot/:userId/toggle",
				"DELETE /api/chatbot/:userId",
			},
		})
	})

	// Session routes
	app.Post("/api/session/init", authMiddleware.Auth, sessionHandler.InitSession)
	app.Get("/api/session/qr/:userId", sessionHandler.GetQRCode)
	app.Get("/api/session/status/:userId", sessionHandler.GetStatus)
	app.Post("/api/session/logout", sessionHandler.Logout)
	app.Get("/api/sessions", sessionHandler.GetAllSessions)

	// Message routes
	app.Post("/api/message/send", authMiddleware.Auth, messageHandler.SendTextMessage)
	app.Post("/api/message/send-many", authMiddleware.Auth, messageHandler.SendBulkTextMessages)
	app.Post("/api/message/send-media", authMiddleware.Auth, messageHandler.SendMediaMessage)
	app.Post("/api/message/send-many-image", authMiddleware.Auth, messageHandler.SendBulkMediaMessages)

	// Chatbot routes
	app.Post("/api/chatbot", authMiddleware.Auth, chatbotHandler.CreateOrUpdateChatbot)
	app.Get("/api/chatbot", authMiddleware.Auth, chatbotHandler.GetChatbot)
	app.Post("/api/chatbot/option", authMiddleware.Auth, chatbotHandler.CreateOrUpdateOption)
	app.Delete("/api/chatbot/option/:userId/:optionKey", chatbotHandler.DeleteOption)
	app.Patch("/api/chatbot/:userId/toggle", chatbotHandler.ToggleChatbot)
	app.Delete("/api/chatbot/:userId", chatbotHandler.DeleteChatbot)

	// Test authenticated endpoint
	app.Get("/yeaboi", authMiddleware.Auth, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data": "INTERNAL POINTER VARIABLE",
		})
	})

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\n🛑 Shutting down gracefully...")

		// Save session metadata before shutdown
		if err := waManager.SaveSessionMetadata(); err != nil {
			log.Printf("Error saving session metadata on shutdown: %v", err)
		}
		app.Shutdown()
	}()

	// Start server
	port := cfg.Server.Port
	log.Printf("\n🚀 WhatsApp API Server running on port %s", port)
	log.Printf("📍 http://localhost:%s", port)
	log.Printf("\n✅ Server ready to accept connections\n")

	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
