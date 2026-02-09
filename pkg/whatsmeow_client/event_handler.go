package whatsmeow_client

import (
	"fmt"

	"go.mau.fi/whatsmeow/types/events"
)

// MessageEvent represents a simplified message event
type MessageEvent struct {
	ID        string
	From      string
	Body      string
	FromMe    bool
	Timestamp int64
	IsGroup   bool
}

// ExtractMessageEvent converts whatsmeow event to simplified MessageEvent
func ExtractMessageEvent(evt interface{}) (*MessageEvent, error) {
	switch v := evt.(type) {
	case *events.Message:
		// Extract message text
		body := ""
		if v.Message.Conversation != nil {
			body = *v.Message.Conversation
		} else if v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.Text != nil {
			body = *v.Message.ExtendedTextMessage.Text
		}

		return &MessageEvent{
			ID:        v.Info.ID,
			From:      v.Info.Chat.String(),
			Body:      body,
			FromMe:    v.Info.IsFromMe,
			Timestamp: v.Info.Timestamp.Unix(),
			IsGroup:   v.Info.IsGroup,
		}, nil

	default:
		return nil, fmt.Errorf("not a message event")
	}
}