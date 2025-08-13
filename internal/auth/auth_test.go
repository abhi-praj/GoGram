package auth

import (
	"testing"
)

func TestNewInstagramAuth(t *testing.T) {
	auth := NewInstagramAuth()
	if auth == nil {
		t.Error("NewInstagramAuth() returned nil")
	}

	if auth.config == nil {
		t.Error("InstagramAuth config is nil")
	}
}

func TestGetCurrentUsername(t *testing.T) {
	auth := NewInstagramAuth()
	username := auth.GetCurrentUsername()
	_ = username
}

func TestIsLoggedIn(t *testing.T) {
	auth := NewInstagramAuth()
	loggedIn := auth.IsLoggedIn()
	if loggedIn {
		t.Error("Expected IsLoggedIn to be false initially")
	}
}
