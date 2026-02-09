package database

import (
	"fmt"
	"log"

	"github.com/yourusername/whatsapp-chatbot-go/internal/config"
	"github.com/yourusername/whatsapp-chatbot-go/internal/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.Server.Env == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("✅ Database connected successfully")

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&domain.User{},
		&domain.Chatbot{},
		&domain.ChatbotOption{},
		&domain.ConversationState{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ Database migrations completed")
	return nil
}