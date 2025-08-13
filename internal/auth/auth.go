package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/abhi-praj/GoGram/internal/client"
	"github.com/abhi-praj/GoGram/internal/config"
)

// InstagramAuth handles Instagram authentication operations
type InstagramAuth struct {
	client *client.ClientWrapper
	config *config.Config
}

// NewInstagramAuth creates a new authentication instance
func NewInstagramAuth() *InstagramAuth {
	return &InstagramAuth{
		config: config.GetInstance(),
	}
}

// Login attempts to login to Instagram, first trying session then username/password
func (a *InstagramAuth) Login() (*client.ClientWrapper, error) {
	// Try to get current username from config
	currentUsername := a.config.Get("login.current_username", "").(string)

	a.client = client.NewClientWrapper(currentUsername)

	// Try to login by session first
	fmt.Println("Attempting to login with saved session...")
	if err := a.client.LoginBySession(); err == nil {
		fmt.Printf("Successfully logged in as @%s\n", a.client.GetUsername())
		return a.client, nil
	}

	// Try by username/password
	fmt.Println("Session login failed, attempting username/password login...")
	return a.loginByUsername()
}

// LoginByUsername prompts for username and password to login
func (a *InstagramAuth) LoginByUsername() (*client.ClientWrapper, error) {
	return a.loginByUsername()
}

// loginByUsername handles the username/password login flow
func (a *InstagramAuth) loginByUsername() (*client.ClientWrapper, error) {
	reader := bufio.NewReader(os.Stdin)

	// Get username
	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read username: %v", err)
	}
	username = strings.TrimSpace(username)

	// Get password
	fmt.Print("Password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %v", err)
	}
	password = strings.TrimSpace(password)

	// Check for 2FA
	var verificationCode string
	fmt.Print("Do you use 2FA? (y/N): ")
	use2FA, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read 2FA response: %v", err)
	}
	use2FA = strings.TrimSpace(strings.ToLower(use2FA))

	if use2FA == "y" || use2FA == "yes" {
		fmt.Print("Verification code (from Auth App): ")
		verificationCode, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read verification code: %v", err)
		}
		verificationCode = strings.TrimSpace(verificationCode)
	}

	// Create client and attempt login
	a.client = client.NewClientWrapper(username)

	fmt.Println("Logging in...")
	if err := a.client.Login(username, password, verificationCode); err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}

	fmt.Printf("Successfully logged in as @%s\n", username)
	return a.client, nil
}

// Logout logs out the current user
func (a *InstagramAuth) Logout(username string) error {
	if username == "" {
		username = a.config.Get("login.current_username", "").(string)
	}

	if username == "" {
		return fmt.Errorf("no username specified and no current username in config")
	}

	// Create client wrapper for the specified username
	client := client.NewClientWrapper(username)

	fmt.Printf("Logging out @%s...\n", username)

	// Try to login by session first to get the client
	if err := client.LoginBySession(); err != nil {
		return fmt.Errorf("failed to load session for @%s: %v", username, err)
	}

	// Logout
	if err := client.Logout(); err != nil {
		return fmt.Errorf("logout failed: %v", err)
	}

	// Clear both current and default username if they match
	if a.config.Get("login.default_username", "").(string) == username {
		a.config.Set("login.default_username", nil)
	}

	fmt.Printf("✅ Successfully logged out @%s\n", username)
	return nil
}

// GetCurrentUsername returns the currently logged in username
func (a *InstagramAuth) GetCurrentUsername() string {
	return a.config.Get("login.current_username", "").(string)
}

// IsLoggedIn checks if there's an active session
func (a *InstagramAuth) IsLoggedIn() bool {
	return a.client != nil && a.client.IsLoggedIn()
}

// GetClient returns the current client wrapper
func (a *InstagramAuth) GetClient() *client.ClientWrapper {
	return a.client
}
