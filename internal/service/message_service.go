package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/yourusername/whatsapp-chatbot-go/internal/utils"
	"github.com/yourusername/whatsapp-chatbot-go/pkg/whatsmeow_client"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

type MessageService struct {
	waManager *whatsmeow_client.Manager
}

func NewMessageService(waManager *whatsmeow_client.Manager) *MessageService {
	return &MessageService{
		waManager: waManager,
	}
}

type SendMessageResponse struct {
	MessageID string `json:"message_id"`
	Timestamp int64  `json:"timestamp"`
}

type BulkSendResult struct {
	Phone   string `json:"phone"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SendTextMessage sends a text message to a single recipient
func (s *MessageService) SendTextMessage(userID, phone, message string) (*SendMessageResponse, error) {
	clientData, exists := s.waManager.GetClient(userID)
	if !exists {
		return nil, fmt.Errorf("WhatsApp session not found. Please initialize session first")
	}

	if clientData.Status != whatsmeow_client.StatusReady {
		return nil, fmt.Errorf("WhatsApp session not ready. Current status: %s", clientData.Status)
	}

	// Format phone number
	jidString := utils.FormatPhoneNumber(phone)
	jid, err := types.ParseJID(jidString)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Send message
	resp, err := clientData.Client.SendMessage(context.Background(), jid, &waProto.Message{
		Conversation: proto.String(message),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &SendMessageResponse{
		MessageID: resp.ID,
		Timestamp: resp.Timestamp.Unix(),
	}, nil
}

// SendMediaMessage sends a media message with caption
func (s *MessageService) SendMediaMessage(userID, phone, imageURL, caption string) (*SendMessageResponse, error) {
	clientData, exists := s.waManager.GetClient(userID)
	if !exists {
		return nil, fmt.Errorf("WhatsApp session not found. Please initialize session first")
	}

	if clientData.Status != whatsmeow_client.StatusReady {
		return nil, fmt.Errorf("WhatsApp session not ready. Current status: %s", clientData.Status)
	}

	// Format phone number
	jidString := utils.FormatPhoneNumber(phone)
	jid, err := types.ParseJID(jidString)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Download media
	httpResp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media: %w", err)
	}

	// Upload media to WhatsApp
	uploaded, err := clientData.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
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

	resp, err := clientData.Client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send media message: %w", err)
	}

	return &SendMessageResponse{
		MessageID: resp.ID,
		Timestamp: resp.Timestamp.Unix(),
	}, nil
}

// SendBulkTextMessages sends text messages to multiple recipients
func (s *MessageService) SendBulkTextMessages(userID string, phones []string, message string) []BulkSendResult {
	clientData, exists := s.waManager.GetClient(userID)
	if !exists {
		// Return error for all phones
		results := make([]BulkSendResult, len(phones))
		for i, phone := range phones {
			results[i] = BulkSendResult{
				Phone:   phone,
				Success: false,
				Error:   "WhatsApp session not found",
			}
		}
		return results
	}

	results := make([]BulkSendResult, len(phones))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, phone := range phones {
		wg.Add(1)
		go func(idx int, ph string) {
			defer wg.Done()

			jidString := utils.FormatPhoneNumber(ph)
			jid, err := types.ParseJID(jidString)
			if err != nil {
				mu.Lock()
				results[idx] = BulkSendResult{
					Phone:   ph,
					Success: false,
					Error:   fmt.Sprintf("invalid phone number: %v", err),
				}
				mu.Unlock()
				return
			}

			_, err = clientData.Client.SendMessage(context.Background(), jid, &waProto.Message{
				Conversation: proto.String(message),
			})

			mu.Lock()
			if err != nil {
				results[idx] = BulkSendResult{
					Phone:   ph,
					Success: false,
					Error:   err.Error(),
				}
			} else {
				results[idx] = BulkSendResult{
					Phone:   ph,
					Success: true,
				}
			}
			mu.Unlock()
		}(i, phone)
	}

	wg.Wait()
	return results
}

// SendBulkMediaMessages sends media messages to multiple recipients
func (s *MessageService) SendBulkMediaMessages(userID string, phones []string, imageURL, message string) []BulkSendResult {
	clientData, exists := s.waManager.GetClient(userID)
	if !exists {
		results := make([]BulkSendResult, len(phones))
		for i, phone := range phones {
			results[i] = BulkSendResult{
				Phone:   phone,
				Success: false,
				Error:   "WhatsApp session not found",
			}
		}
		return results
	}

	// Download and upload media once
	httpResp, err := http.Get(imageURL)
	if err != nil {
		results := make([]BulkSendResult, len(phones))
		for i, phone := range phones {
			results[i] = BulkSendResult{
				Phone:   phone,
				Success: false,
				Error:   fmt.Sprintf("failed to download media: %v", err),
			}
		}
		return results
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		results := make([]BulkSendResult, len(phones))
		for i, phone := range phones {
			results[i] = BulkSendResult{
				Phone:   phone,
				Success: false,
				Error:   fmt.Sprintf("failed to read media: %v", err),
			}
		}
		return results
	}

	uploaded, err := clientData.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err != nil {
		results := make([]BulkSendResult, len(phones))
		for i, phone := range phones {
			results[i] = BulkSendResult{
				Phone:   phone,
				Success: false,
				Error:   fmt.Sprintf("failed to upload media: %v", err),
			}
		}
		return results
	}

	results := make([]BulkSendResult, len(phones))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, phone := range phones {
		wg.Add(1)
		go func(idx int, ph string) {
			defer wg.Done()

			jidString := utils.FormatPhoneNumber(ph)
			jid, err := types.ParseJID(jidString)
			if err != nil {
				mu.Lock()
				results[idx] = BulkSendResult{
					Phone:   ph,
					Success: false,
					Error:   fmt.Sprintf("invalid phone number: %v", err),
				}
				mu.Unlock()
				return
			}

			msg := &waProto.Message{
				ImageMessage: &waProto.ImageMessage{
					Caption:       proto.String(message),
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

			mu.Lock()
			if err != nil {
				results[idx] = BulkSendResult{
					Phone:   ph,
					Success: false,
					Error:   err.Error(),
				}
			} else {
				results[idx] = BulkSendResult{
					Phone:   ph,
					Success: true,
				}
			}
			mu.Unlock()
		}(i, phone)
	}

	wg.Wait()
	return results
}