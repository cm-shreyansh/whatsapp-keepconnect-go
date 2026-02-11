package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateID generates a unique ID with optional prefix
func GenerateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return prefix + hex.EncodeToString(bytes)
}

// IsGreeting checks if a message is a greeting
func IsGreeting(message string) bool {
	greetings := []string{"hi", "hello", "hey", "hii", "hiii", "hiiii", "hiiii", "helo", "hola", "help"}
	normalized := strings.ToLower(strings.TrimSpace(message))

	for _, greeting := range greetings {
		if normalized == greeting {
			return true
		}
	}

	return false
}

// FormatPhoneNumber formats a phone number to WhatsApp JID format
func FormatPhoneNumber(phone string) string {
	// Remove all non-digit characters
	digits := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digits += string(char)
		}
	}

	// If it's a 10-digit number, assume it's Indian and prepend 91
	if len(digits) == 10 {
		digits = "91" + digits
	}

	// Return in WhatsApp format
	return digits + "@s.whatsapp.net"
}

// PtrString returns a pointer to a string
func PtrString(s string) *string {
	return &s
}

// PtrInt returns a pointer to an int
func PtrInt(i int) *int {
	return &i
}

// SafeString safely dereferences a string pointer
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ValidateRequired checks if required fields are present
func ValidateRequired(fields map[string]string) error {
	missing := []string{}

	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}
