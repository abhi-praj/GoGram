package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Davincible/goinsta/v3"
	"github.com/abhi-praj/GoGram/internal/config"
)

// ClientWrapper wraps the goinsta Instagram client with my additional functionality
type ClientWrapper struct {
	instaClient *goinsta.Instagram
	username    string
	config      *config.Config
}

// NewClientWrapper creates a new client wrapper
func NewClientWrapper(username string) *ClientWrapper {
	return &ClientWrapper{
		username: username,
		config:   config.GetInstance(),
	}
}

// Login attempts to login using saved session, falls back to username/password
func (c *ClientWrapper) Login(username, password string, verificationCode string) error {
	c.instaClient = goinsta.New(username, password)

	// reminder to check how 2fa works

	// Attempt to login
	if err := c.instaClient.Login(); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	// Update username and save session
	c.username = username
	c.config.Set("login.current_username", username)
	c.config.Set("login.default_username", username)

	// Save session
	return c.saveSession()
}

// LoginBySession attempts to login using a saved session
func (c *ClientWrapper) LoginBySession() error {
	if c.username == "" {
		// Try to get username from config
		if username := c.config.Get("login.current_username", ""); username != "" {
			c.username = username.(string)
		} else {
			return fmt.Errorf("no username specified and no current username in config")
		}
	}

	// Try to import session
	sessionPath := c.getSessionPath()
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("no session file found for user %s", c.username)
	}

	var err error
	c.instaClient, err = goinsta.Import(sessionPath)
	if err != nil {
		return fmt.Errorf("failed to import session: %v", err)
	}

	return nil
}

// Logout logs out from Instagram and clears session
func (c *ClientWrapper) Logout() error {
	if c.instaClient != nil {
		if err := c.instaClient.Logout(); err != nil {
			return fmt.Errorf("logout failed: %v", err)
		}
	}

	// Clear session and username
	sessionPath := c.getSessionPath()
	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %v", err)
	}

	c.config.Set("login.current_username", nil)
	c.instaClient = nil

	return nil
}

// saveSession saves the current session
func (c *ClientWrapper) saveSession() error {
	if c.instaClient == nil {
		return fmt.Errorf("no active Instagram client")
	}

	sessionPath := c.getSessionPath()

	if err := os.MkdirAll(filepath.Dir(sessionPath), 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %v", err)
	}

	if err := c.instaClient.Export(sessionPath); err != nil {
		return fmt.Errorf("failed to export session: %v", err)
	}

	return nil
}

// getSessionPath returns the path to the session file
func (c *ClientWrapper) getSessionPath() string {
	usersDir := c.config.Get("advanced.users_dir", "").(string)
	return filepath.Join(usersDir, c.username, "session.json")
}

// GetUsername returns the current username
func (c *ClientWrapper) GetUsername() string {
	return c.username
}

// GetInstaClient returns the underlying goinsta client
func (c *ClientWrapper) GetInstaClient() *goinsta.Instagram {
	return c.instaClient
}

// IsLoggedIn checks if the client is currently logged in
func (c *ClientWrapper) IsLoggedIn() bool {
	return c.instaClient != nil
}

// RefreshSession refreshes the current session
func (c *ClientWrapper) RefreshSession() error {
	if c.instaClient == nil {
		return fmt.Errorf("no active Instagram client")
	}

	// Save current session
	return c.saveSession()
}
