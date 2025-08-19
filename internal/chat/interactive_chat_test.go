package chat

import (
	"testing"
	"time"
)

func TestNewInteractiveChat(t *testing.T) {
	// Mock DirectMessages instance
	dm := &DirectMessages{}

	ic := NewInteractiveChat(dm, "123456")

	if ic.dm != dm {
		t.Errorf("Expected dm to be %v, got %v", dm, ic.dm)
	}

	if ic.chatID != "123456" {
		t.Errorf("Expected chatID to be '123456', got %s", ic.chatID)
	}

	if ic.stopChan == nil {
		t.Error("Expected stopChan to be initialized")
	}
}

func TestIsSubcommand(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"list", true},
		{"send", false},
		{"history", false},
		{"search", false},
		{"123456", false},
		{"abc123", false},
		{"LIST", true},
		{"Send", false},
	}

	for _, tc := range testCases {
		result := IsSubcommand(tc.input)
		if result != tc.expected {
			t.Errorf("IsSubcommand(%s) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestDisplayMessage(t *testing.T) {
	ic := &InteractiveChat{}

	msg := &Message{
		Text:      "Hello world",
		Sender:    "TestUser",
		Timestamp: time.Now(),
	}

	// This should not panic
	ic.displayMessage(msg, false)
	ic.displayMessage(msg, true)
}
