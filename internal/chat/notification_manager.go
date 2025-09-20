package chat

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Davincible/goinsta/v3"
)

// NotificationManager handles background message notifications
type NotificationManager struct {
	dm             *DirectMessages
	lastMessageIDs map[string]string
	lastCheckTimes map[string]time.Time
	mutex          sync.Mutex
	stopChan       chan bool
	isRunning      bool
	checkInterval  time.Duration
	isPaused       bool
	pauseMutex     sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(dm *DirectMessages) *NotificationManager {
	return &NotificationManager{
		dm:             dm,
		lastMessageIDs: make(map[string]string),
		lastCheckTimes: make(map[string]time.Time),
		stopChan:       make(chan bool),
		checkInterval:  5 * time.Second,
		isPaused:       false,
	}
}

// Start begins background message monitoring
func (nm *NotificationManager) Start() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.isRunning {
		return fmt.Errorf("notification manager already running")
	}

	if nm.dm == nil || nm.dm.insta == nil {
		return fmt.Errorf("not logged in")
	}

	nm.isRunning = true
	// Initialize last check times
	nm.initializeLastCheckTimes()

	// Start background monitoring
	go nm.backgroundMonitor()
	return nil
}

// Stop stops background message monitoring
func (nm *NotificationManager) Stop() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.isRunning {
		nm.isRunning = false
		close(nm.stopChan)
	}
}

// Pause temporarily pauses notifications (e.g., during interactive chat)
func (nm *NotificationManager) Pause() {
	nm.pauseMutex.Lock()
	defer nm.pauseMutex.Unlock()
	nm.isPaused = true
}

// Resume resumes notifications after being paused
func (nm *NotificationManager) Resume() {
	nm.pauseMutex.Lock()
	defer nm.pauseMutex.Unlock()
	nm.isPaused = false
}

// IsPaused returns whether notifications are currently paused
func (nm *NotificationManager) IsPaused() bool {
	nm.pauseMutex.RLock()
	defer nm.pauseMutex.RUnlock()
	return nm.isPaused
}

// initializeLastCheckTimes sets up initial state for all chats
func (nm *NotificationManager) initializeLastCheckTimes() {
	chats, err := nm.dm.GetChats()
	if err != nil {
		return
	}

	now := time.Now()
	for _, chat := range chats {
		nm.lastCheckTimes[chat.InternalID] = now
		// Get the last message ID if available
		messages, err := nm.dm.GetChatHistory(chat.InternalID, 1)
		if err == nil && len(messages) > 0 {
			nm.lastMessageIDs[chat.InternalID] = messages[0].ID
		}
	}
}

// Refresh initializes the notification system with current chat state
func (nm *NotificationManager) Refresh() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Clear existing state
	nm.lastMessageIDs = make(map[string]string)
	nm.lastCheckTimes = make(map[string]time.Time)

	// Re-initialize
	nm.initializeLastCheckTimes()
}

// backgroundMonitor runs in the background and checks for new messages
func (nm *NotificationManager) backgroundMonitor() {
	ticker := time.NewTicker(nm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.checkForNewMessages()
		}
	}
}

// checkForNewMessages checks all chats for new messages
func (nm *NotificationManager) checkForNewMessages() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if !nm.isRunning {
		return
	}

	// Check if notifications are paused
	if nm.IsPaused() {
		return
	}

	chats, err := nm.dm.GetChats()
	if err != nil {
		return
	}

	for _, chat := range chats {
		nm.checkChatForNewMessages(chat)
	}
}

// checkChatForNewMessages checks a specific chat for new messages
func (nm *NotificationManager) checkChatForNewMessages(chat *Chat) {
	// Sync inbox to get latest data
	if err := nm.dm.insta.Inbox.Sync(); err != nil {
		return
	}

	// Find the conversation
	var conversation *goinsta.Conversation
	for _, conv := range nm.dm.insta.Inbox.Conversations {
		if conv.ID == chat.ID {
			conversation = conv
			break
		}
	}

	if conversation == nil {
		return
	}

	// Get latest items
	if err := conversation.GetItems(); err != nil {
		return
	}

	lastMessageID := nm.lastMessageIDs[chat.InternalID]
	lastCheckTime := nm.lastCheckTimes[chat.InternalID]

	// Check for new messages (messages newer than our last check)
	if len(conversation.Items) > 0 {
		latestItem := conversation.Items[0]
		latestTime := time.Unix(latestItem.Timestamp, 0)

		// Skip if this is the same message we've already seen
		if latestItem.ID == lastMessageID {
			return
		}

		// Skip messages sent by current user
		currentUserIDInt, _ := strconv.ParseInt(nm.dm.currentUserID, 10, 64)
		if latestItem.UserID == currentUserIDInt {
			return
		}

		// Only show messages that are genuinely new (within last 30 seconds of last check)
		// This prevents showing old messages that might be loaded from cache
		timeSinceLastCheck := time.Since(lastCheckTime)
		if latestTime.After(lastCheckTime) && timeSinceLastCheck < 30*time.Second {
			// Create message object
			msg := &Message{
				ID:        latestItem.ID,
				Text:      latestItem.Text,
				Timestamp: latestTime,
				Type:      "text",
			}

			// Determine sender name from conversation users
			var senderName string
			for _, user := range conversation.Users {
				if user.ID == latestItem.UserID {
					if user.FullName != "" {
						senderName = user.FullName
					} else {
						senderName = user.Username
					}
					break
				}
			}

			if senderName == "" {
				senderName = "Unknown User"
			}
			msg.Sender = senderName

			nm.displayNotification(chat, msg)
		}
	}

	// Update tracking info
	if len(conversation.Items) > 0 {
		nm.lastMessageIDs[chat.InternalID] = conversation.Items[0].ID
	}
	nm.lastCheckTimes[chat.InternalID] = time.Now()
}

// displayNotification shows a notification for a new message
func (nm *NotificationManager) displayNotification(chat *Chat, msg *Message) {
	// Get sender display name
	senderDisplay := msg.Sender
	if senderDisplay == "Unknown User" {
		// Try to get username from chat users as fallback
		for _, user := range chat.Users {
			if user.FullName == msg.Sender || user.Username == msg.Sender {
				senderDisplay = user.Username
				break
			}
		}
	}

	// Truncate message for preview
	preview := msg.Text
	if len(preview) > 50 {
		preview = preview[:47] + "..."
	}

	// Display notification with timestamp - make it stand out
	timeStr := msg.Timestamp.Format("15:04")
	fmt.Printf("\n" + strings.Repeat("─", 60) + "\n")
	fmt.Printf("🔔 [%s] New message from %s in %s\n",
		timeStr, senderDisplay, chat.Title)
	fmt.Printf("💬 %s\n", preview)
	fmt.Printf("💬 Use 'chat %s' to open this conversation\n", chat.InternalID)
	fmt.Printf(strings.Repeat("─", 60) + "\n")
	fmt.Print("ig-cli> ")
}

// GetDebugInfo returns debug information about the notification system
func (nm *NotificationManager) GetDebugInfo() map[string]interface{} {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	info := make(map[string]interface{})
	info["isRunning"] = nm.isRunning
	info["isPaused"] = nm.IsPaused()
	info["checkInterval"] = nm.checkInterval
	info["lastMessageIDs"] = nm.lastMessageIDs
	info["lastCheckTimes"] = nm.lastCheckTimes

	return info
}

// IsRunning returns whether the notification manager is active
func (nm *NotificationManager) IsRunning() bool {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	return nm.isRunning
}
