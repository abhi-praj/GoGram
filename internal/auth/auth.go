package auth

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/abhi-praj/GoGram/internal/config"
	"github.com/abhi-praj/GoGram/internal/client"
)

type AuthManager struct {
	config *config.Config
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		config: config.GetInstance(),
	}
}

func (am *AuthManager) Login() (*client.ClientWrapper, error) {
	currentUsername := am.config.GetString("login.current_username", "")
	
	if currentUsername == "" {
		return am.loginByUsername()
	}

	client, err := am.loginBySession(currentUsername)
	if err != nil {
		fmt.Printf("Cannot log in via session: %v\n", err)
		fmt.Println("Logging in with username and password.")
		return am.loginByUsername()
	}

	fmt.Printf("Logged in as %s\n", client.GetUsername())
	return client, nil
}

func (am *AuthManager) loginBySession(username string) (*client.ClientWrapper, error) {
	client := client.NewClientWrapper(username)
	
	if err := client.LoginBySession(); err != nil {
		return nil, err
	}
	
	return client, nil
}

func (am *AuthManager) loginByUsername() (*client.ClientWrapper, error) {
	username := am.promptUsername()
	password := am.promptPassword()
	
	var verificationCode string
	if am.prompt2FA() {
		verificationCode = am.promptVerificationCode()
	}

	client := client.NewClientWrapper(username)
	
	if err := client.Login(username, password, verificationCode); err != nil {
		return nil, fmt.Errorf("error logging in: %v", err)
	}

	am.config.Set("login.current_username", username)
	if am.config.GetString("login.default_username", "") == "" {
		am.config.Set("login.default_username", username)
	}

	fmt.Printf("Logged in as %s\n", username)
	return client, nil
}

func (am *AuthManager) Logout(username string) error {
	if username == "" {
		username = am.config.GetString("login.current_username", "")
	}

	if username == "" {
		fmt.Println("No active session found.")
		return nil
	}

	client := client.NewClientWrapper(username)
	
	if err := client.LoginBySession(); err != nil {
		fmt.Printf("@%s not logged in.\n", username)
		return nil
	}

	if err := client.Logout(); err != nil {
		return fmt.Errorf("error logging out: %v", err)
	}

	if am.config.GetString("login.default_username", "") == username {
		am.config.Set("login.default_username", "")
	}
	
	if am.config.GetString("login.current_username", "") == username {
		am.config.Set("login.current_username", "")
	}

	fmt.Printf("Logged out @%s.\n", username)
	return nil
}

func (am *AuthManager) SwitchAccount(username string) error {
	sessionPath := filepath.Join(
		am.config.GetString("advanced.users_dir", ""),
		username,
		"session.json",
	)

	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("cannot switch to @%s. No session found.\nTry logging in with @%s first.", username, username)
	}

	am.config.Set("login.current_username", username)
	fmt.Printf("Switched to @%s\n", username)
	return nil
}

func (am *AuthManager) promptUsername() string {
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	return strings.TrimSpace(username)
}

func (am *AuthManager) promptPassword() string {
	fmt.Print("Password: ")
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password)
}

func (am *AuthManager) prompt2FA() bool {
	fmt.Print("Do you use 2FA (2 Factor Authentication)? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func (am *AuthManager) promptVerificationCode() string {
	fmt.Print("Provide your verification code (From The Auth App, SMS not supported): ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	return strings.TrimSpace(code)
}

func (am *AuthManager) GetCurrentUsername() string {
	return am.config.GetString("login.current_username", "")
}

func (am *AuthManager) GetDefaultUsername() string {
	return am.config.GetString("login.default_username", "")
}

func (am *AuthManager) IsLoggedIn() bool {
	username := am.GetCurrentUsername()
	if username == "" {
		return false
	}

	sessionPath := filepath.Join(
		am.config.GetString("advanced.users_dir", ""),
		username,
		"session.json",
	)

	_, err := os.Stat(sessionPath)
	return err == nil
}

func (am *AuthManager) ListAccounts() ([]string, error) {
	usersDir := am.config.GetString("advanced.users_dir", "")
	
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, err
	}

	var accounts []string
	for _, entry := range entries {
		if entry.IsDir() {
			sessionPath := filepath.Join(usersDir, entry.Name(), "session.json")
			if _, err := os.Stat(sessionPath); err == nil {
				accounts = append(accounts, entry.Name())
			}
		}
	}

	return accounts, nil
}
