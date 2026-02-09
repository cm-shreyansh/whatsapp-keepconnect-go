package whatsmeow_client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SessionStatus string

const (
	StatusInitializing   SessionStatus = "initializing"
	StatusQRReady        SessionStatus = "qr_ready"
	StatusAuthenticated  SessionStatus = "authenticated"
	StatusReady          SessionStatus = "ready"
	StatusAuthFailed     SessionStatus = "auth_failed"
	StatusDisconnected   SessionStatus = "disconnected"
	StatusNotInitialized SessionStatus = "not_initialized"
)

type ClientData struct {
	Client    *whatsmeow.Client
	status    SessionStatus
	qrCode    string
	Container *sqlstore.Container
	mu        sync.RWMutex
}

// GetStatus safely returns the current status
func (cd *ClientData) GetStatus() SessionStatus {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return cd.status
}

// SetStatus safely sets the status
func (cd *ClientData) SetStatus(status SessionStatus) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.status = status
}

// GetQRCode safely returns the QR code
func (cd *ClientData) GetQRCode() string {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	return cd.qrCode
}

// SetQRCode safely sets the QR code
func (cd *ClientData) SetQRCode(qr string) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.qrCode = qr
}

type Manager struct {
	clients      map[string]*ClientData
	container    *sqlstore.Container
	eventHandler EventHandler
	mu           sync.RWMutex
}

type EventHandler interface {
	HandleMessage(userID string, message interface{})
}

// func NewManager(dbPath string, handler EventHandler) (*Manager, error) {
// 	// Ensure sessions directory exists
// 	if err := os.MkdirAll("./sessions", 0755); err != nil {
// 		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
// 	}

// 	// Create SQLite database for whatsmeow with foreign keys enabled in DSN
// 	dsn := dbPath + "?_foreign_keys=on"
// 	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open whatsmeow database: %w", err)
// 	}

// 	sqlDB, err := db.DB()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get database instance: %w", err)
// 	}

// 	// Create store container
// 	container := sqlstore.NewWithDB(sqlDB, "sqlite3", waLog.Noop)
// 	if err := container.Upgrade(context.Background()); err != nil {
// 		return nil, fmt.Errorf("failed to upgrade store: %w", err)
// 	}

// 	return &Manager{
// 		clients:      make(map[string]*ClientData),
// 		container:    container,
// 		eventHandler: handler,
// 	}, nil
// }
// Update NewManager in pkg/whatsmeow_client/client.go

func NewManager(dbPath string, handler EventHandler) (*Manager, error) {
	// Ensure sessions directory exists
	if err := os.MkdirAll("./sessions", 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// Create SQLite database for whatsmeow with foreign keys enabled in DSN
	dsn := dbPath + "?_foreign_keys=on"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open whatsmeow database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Create store container
	container := sqlstore.NewWithDB(sqlDB, "sqlite3", waLog.Noop)
	if err := container.Upgrade(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to upgrade store: %w", err)
	}

	manager := &Manager{
		clients:      make(map[string]*ClientData),
		container:    container,
		eventHandler: handler,
	}

	// Auto-restore previous sessions
	if err := manager.RestoreSessions(); err != nil {
		log.Printf("Warning: failed to restore sessions: %v", err)
		// Don't fail initialization, just log the warning
	}

	return manager, nil
}

func (m *Manager) GetOrCreateClient(userID string) (*ClientData, error) {
	fmt.Print("HERE I AM")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if client already exists
	foundClientData, exists := m.clients[userID]
	fmt.Print(foundClientData)
	if exists {
		fmt.Sprint("%v", exists)
		return foundClientData, nil
	}

	// Get device from store
	device, err := m.container.GetFirstDevice(context.Background())
	if err != nil {
		// Create new device if none exists
		device = m.container.NewDevice()
	}

	// Create WhatsApp client
	client := whatsmeow.NewClient(device, waLog.Noop)

	fmt.Print("HERE WE ARE")
	fmt.Print(client.IsLoggedIn())
	fmt.Print(client.IsConnected())
	clientData := &ClientData{
		Client:    client,
		Container: m.container,
	}
	clientData.SetStatus(StatusInitializing)

	// Set up event handlers
	m.setupEventHandlers(userID, clientData)

	m.clients[userID] = clientData
	clientBro, exists := m.clients[userID]
	fmt.Print("WE're set yeaa")
	fmt.Print(exists)
	fmt.Print(m.clients[userID].status)
	fmt.Print("MAIN THING DUDE")
	fmt.Print(clientBro)
	return clientData, nil
}

func (m *Manager) setupEventHandlers(userID string, clientData *ClientData) {
	client := clientData.Client

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.LoggedOut:
			clientData.SetStatus(StatusDisconnected)

		case *events.Connected:
			clientData.SetStatus(StatusReady)

		case *events.Disconnected:
			clientData.SetStatus(StatusDisconnected)

		default:
			// Pass message events to the event handler
			if m.eventHandler != nil {
				m.eventHandler.HandleMessage(userID, v)
			}
		}
	})
}

func (m *Manager) InitializeClient(userID string) (*ClientData, error) {
	fmt.Print("HERE I AM")
	clientData, err := m.GetOrCreateClient(userID)
	if err != nil {
		return nil, err
	}

	if clientData.Client.Store.ID == nil {
		// Not logged in, generate QR code
		qrChan, err := clientData.Client.GetQRChannel(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get QR channel: %w", err)
		}

		// Connect to WhatsApp
		if err := clientData.Client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}

		// Wait for QR code
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					// Generate QR code as base64 data URL
					png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
					if err != nil {
						fmt.Printf("Failed to generate QR code: %v\n", err)
						continue
					}

					base64Str := base64.StdEncoding.EncodeToString(png)
					qrDataURL := "data:image/png;base64," + base64Str

					clientData.SetQRCode(qrDataURL)
					clientData.SetStatus(StatusQRReady)
				} else {
					// QR code scanned or error
					clientData.SetStatus(StatusAuthenticated)
				}
			}
		}()
	} else {
		// Already logged in, just connect
		if err := clientData.Client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}
		clientData.SetStatus(StatusReady)
	}

	return clientData, nil
}

func (m *Manager) GetClient(userID string) (*ClientData, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clientData, exists := m.clients[userID]
	fmt.Print(exists)
	// fmt.Print(m.clients[userID].status)
	return clientData, exists
}

func (m *Manager) LogoutClient(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	clientData, exists := m.clients[userID]
	if !exists {
		return fmt.Errorf("client not found")
	}

	// 1. Logout from WhatsApp (deletes device from WhatsApp servers)
	if err := clientData.Client.Logout(context.Background()); err != nil {
		log.Printf("Warning: logout error for %s: %v", userID, err)
		// Continue cleanup even if logout fails
	}

	// 2. Disconnect the client
	clientData.Client.Disconnect()

	// 3. Remove from memory
	delete(m.clients, userID)

	// 4. Update metadata file
	if err := m.SaveSessionMetadata(); err != nil {
		log.Printf("Warning: failed to save metadata after logout: %v", err)
	}

	log.Printf("✅ Session cleaned up for user %s", userID)
	return nil
}

func (m *Manager) GetAllSessions() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]map[string]interface{}, 0, len(m.clients))
	for userID, clientData := range m.clients {
		status := clientData.GetStatus()
		sessions = append(sessions, map[string]interface{}{
			"user_id":      userID,
			"status":       status,
			"is_logged_in": status == StatusReady,
		})
	}

	return sessions
}

// SendTextMessage sends a text message to a phone number
func (cd *ClientData) SendTextMessage(phone, message string) (string, error) {
	jid, err := types.ParseJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	resp, err := cd.Client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: &message,
	})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// SendMediaMessage sends a media message with caption
func (cd *ClientData) SendMediaMessage(phone, mediaURL, caption string) (string, error) {
	jid, err := types.ParseJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %w", err)
	}

	// Download media
	resp, err := http.Get(mediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read media: %w", err)
	}

	// Upload media to WhatsApp
	uploaded, err := cd.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	// Send image message
	msg := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(data)),
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(data))),
		},
	}

	sendResp, err := cd.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return "", err
	}

	return sendResp.ID, nil
}

type SessionMetadata struct {
	UserID       string    `json:"user_id"`
	LastActivity time.Time `json:"last_activity"`
	Status       string    `json:"status"`
}

// SaveSessionMetadata saves active session info to disk
func (m *Manager) SaveSessionMetadata() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metadata := make([]SessionMetadata, 0, len(m.clients))
	for userID, clientData := range m.clients {
		metadata = append(metadata, SessionMetadata{
			UserID:       userID,
			LastActivity: time.Now(),
			Status:       string(clientData.GetStatus()),
		})
	}

	// Ensure directory exists
	dir := filepath.Dir("./sessions/metadata.json")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("./sessions/metadata.json", data, 0644)
}

// LoadSessionMetadata loads session info from disk
func loadSessionMetadata() ([]SessionMetadata, error) {
	data, err := os.ReadFile("./sessions/metadata.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata file yet
		}
		return nil, err
	}

	var metadata []SessionMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// RestoreSessions automatically restores all previously active sessions
func (m *Manager) RestoreSessions() error {
	metadata, err := loadSessionMetadata()
	if err != nil {
		return fmt.Errorf("failed to load session metadata: %w", err)
	}

	if metadata == nil {
		log.Println("No previous sessions to restore")
		return nil
	}

	log.Printf("Restoring %d previous sessions...", len(metadata))

	// Get all devices from whatsmeow store
	devices, err := m.container.GetAllDevices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	// Create a map of device IDs for quick lookup
	deviceMap := make(map[string]*store.Device)
	for _, device := range devices {
		if device.ID != nil {
			// Use the JID as the user identifier
			userID := device.ID.User
			deviceMap[userID] = device
		}
	}

	restored := 0
	for _, meta := range metadata {
		// Check if device still exists in store
		device, exists := deviceMap[meta.UserID]
		if !exists {
			log.Printf("Skipping session for %s - device not found in store", meta.UserID)
			continue
		}

		// Create WhatsApp client with existing device
		client := whatsmeow.NewClient(device, waLog.Noop)

		clientData := &ClientData{
			Client:    client,
			Container: m.container,
		}
		clientData.SetStatus(StatusInitializing)

		// Set up event handlers
		m.setupEventHandlers(meta.UserID, clientData)

		// Add to clients map
		m.mu.Lock()
		m.clients[meta.UserID] = clientData
		m.mu.Unlock()

		// Connect to WhatsApp
		if err := client.Connect(); err != nil {
			log.Printf("Failed to restore session for %s: %v", meta.UserID, err)
			continue
		}

		restored++
		log.Printf("✅ Restored session for user %s", meta.UserID)
	}

	log.Printf("Successfully restored %d/%d sessions", restored, len(metadata))
	return nil
}

// StartMetadataSaver starts a background goroutine to periodically save metadata
func (m *Manager) StartMetadataSaver(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if err := m.SaveSessionMetadata(); err != nil {
				log.Printf("Error saving session metadata: %v", err)
			}
		}
	}()
}
