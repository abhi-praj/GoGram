package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abhi-praj/GoGram/internal/config"
	"github.com/Davincible/goinsta/v3"
)

type ClientWrapper struct {
	username     string
	instaClient  *goinsta.Insta
	sessionPath  string
	config       *config.Config
}

func NewClientWrapper(username string) *ClientWrapper {
	cfg := config.GetInstance()
	
	if username == "" {
		username = cfg.GetString("login.current_username", "")
		if username == "" {
			username = cfg.GetString("login.default_username", "")
		}
	}

	sessionPath := filepath.Join(
		cfg.GetString("advanced.users_dir", ""),
		username,
		"session.json",
	)

	return &ClientWrapper{
		username:    username,
		sessionPath: sessionPath,
		config:      cfg,
	}
}

func (cw *ClientWrapper) Login(username, password, verificationCode string) error {
	insta := goinsta.New(username, password)
	
	insta.SetUserAgent("Instagram 219.0.0.12.117 Android")
	
	if verificationCode != "" {
		insta.TwoFactorInfo = &goinsta.TwoFactorInfo{
			Username: username,
		}
		insta.ChallengeCode = verificationCode
	}

	if err := insta.Login(); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	cw.instaClient = insta
	if err := cw.saveSession(); err != nil {
		return fmt.Errorf("failed to save session: %v", err)
	}

	return nil
}

func (cw *ClientWrapper) LoginBySession() error {
	if cw.username == "" {
		return fmt.Errorf("username not set")
	}

	if _, err := os.Stat(cw.sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("session file not found for user %s", cw.username)
	}

	insta, err := goinsta.Import(cw.sessionPath)
	if err != nil {
		return fmt.Errorf("failed to import session: %v", err)
	}

	if err := insta.OpenApp(); err != nil {
		cw.deleteSession()
		return fmt.Errorf("session expired: %v", err)
	}

	cw.instaClient = insta
	return nil
}

func (cw *ClientWrapper) Logout() error {
	if cw.instaClient == nil {
		return fmt.Errorf("not logged in")
	}

	if err := cw.instaClient.Logout(); err != nil {
		return fmt.Errorf("logout failed: %v", err)
	}

	cw.deleteSession()
	
	cw.instaClient = nil

	return nil
}

func (cw *ClientWrapper) saveSession() error {
	if cw.instaClient == nil {
		return fmt.Errorf("no active session to save")
	}

	sessionDir := filepath.Dir(cw.sessionPath)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %v", err)
	}

	if err := cw.instaClient.Export(cw.sessionPath); err != nil {
		return fmt.Errorf("failed to export session: %v", err)
	}

	return nil
}

func (cw *ClientWrapper) deleteSession() error {
	if _, err := os.Stat(cw.sessionPath); err == nil {
		return os.Remove(cw.sessionPath)
	}
	return nil
}

func (cw *ClientWrapper) GetUsername() string {
	return cw.username
}

func (cw *ClientWrapper) GetInstaClient() *goinsta.Insta {
	return cw.instaClient
}

func (cw *ClientWrapper) IsLoggedIn() bool {
	return cw.instaClient != nil
}

func (cw *ClientWrapper) RefreshSession() error {
	if cw.instaClient == nil {
		return fmt.Errorf("not logged in")
	}

	if err := cw.instaClient.OpenApp(); err != nil {
		return fmt.Errorf("failed to refresh session: %v", err)
	}

	return cw.saveSession()
}

func (cw *ClientWrapper) GetUserInfo() (*goinsta.User, error) {
	if cw.instaClient == nil {
		return nil, fmt.Errorf("not logged in")
	}

	return cw.instaClient.Account, nil
}

func (cw *ClientWrapper) GetUserID() int64 {
	if cw.instaClient == nil || cw.instaClient.Account == nil {
		return 0
	}
	return cw.instaClient.Account.ID
}

func (cw *ClientWrapper) GetSessionPath() string {
	return cw.sessionPath
}

func (cw *ClientWrapper) Cleanup(all bool) error {
	if all {
		usersDir := cw.config.GetString("advanced.users_dir", "")
		if usersDir != "" {
			if err := os.RemoveAll(usersDir); err != nil {
				return fmt.Errorf("failed to clean users directory: %v", err)
			}
		}

		cacheDir := cw.config.GetString("advanced.cache_dir", "")
		if cacheDir != "" {
			if err := os.RemoveAll(cacheDir); err != nil {
				return fmt.Errorf("failed to clean cache directory: %v", err)
			}
		}

		mediaDir := cw.config.GetString("advanced.media_dir", "")
		if mediaDir != "" {
			if err := os.RemoveAll(mediaDir); err != nil {
				return fmt.Errorf("failed to clean media directory: %v", err)
			}
		}

		generatedDir := cw.config.GetString("advanced.generated_dir", "")
		if generatedDir != "" {
			if err := os.RemoveAll(generatedDir); err != nil {
				return fmt.Errorf("failed to clean generated directory: %v", err)
			}
		}
	} else {
		cw.deleteSession()
	}

	return nil
}

func (cw *ClientWrapper) GetSessionInfo() (*SessionInfo, error) {
	if cw.instaClient == nil {
		return nil, fmt.Errorf("not logged in")
	}

	fileInfo, err := os.Stat(cw.sessionPath)
	if err != nil {
		return nil, err
	}

	return &SessionInfo{
		Username:    cw.username,
		SessionPath: cw.sessionPath,
		CreatedAt:   fileInfo.ModTime(),
		UserID:      cw.GetUserID(),
	}, nil
}

type SessionInfo struct {
	Username    string    `json:"username"`
	SessionPath string    `json:"session_path"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      int64     `json:"user_id"`
}

func (si *SessionInfo) String() string {
	data, _ := json.MarshalIndent(si, "", "  ")
	return string(data)
}
