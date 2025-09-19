package chat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"strconv"

	"github.com/Davincible/goinsta/v3"
	"github.com/abhi-praj/ig-tui/internal/client"
)

// DirectMessages handles Instagram direct messaging functionality
type DirectMessages struct {
	client         *client.ClientWrapper
	insta          *goinsta.Instagram
	internalIDMap  map[string]string
	nextInternalID int
	currentUserID  string
}

// NewDirectMessages creates a new DirectMessages instance
func NewDirectMessages(client *client.ClientWrapper) *DirectMessages {
	dm := &DirectMessages{
		client:         client,
		insta:          client.GetInstaClient(),
		internalIDMap:  make(map[string]string),
		nextInternalID: 100000,
		currentUserID:  client.GetUserID(),
	}
	return dm
}

// Chat represents a single chat conversation
type Chat struct {
	ID           string
	InternalID   string
	Title        string
	Users        []*goinsta.User
	LastMessage  string
	LastActivity time.Time
	UnreadCount  int
	IsGroup      bool
}

// Message represents a single message in a chat
type Message struct {
	ID        string
	Text      string
	Sender    string
	Timestamp time.Time
	Type      string // text, media, etc.
}

// GetChats fetches the list of recent chats
func (dm *DirectMessages) GetChats() ([]*Chat, error) {
	return dm.GetChatsWithLimit(5)
}

// GetChatsWithLimit fetches the list of recent chats with a limit
func (dm *DirectMessages) GetChatsWithLimit(limit int) ([]*Chat, error) {
	if dm.insta == nil {
		return nil, fmt.Errorf("not logged in")
	}

	// Sync inbox to get latest data
	if err := dm.insta.Inbox.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync inbox: %v", err)
	}

	var chats []*Chat
	conversations := dm.insta.Inbox.Conversations

	// my attempt to sort by last activity at
	sortableConvs := make([]*goinsta.Conversation, len(conversations))
	copy(sortableConvs, conversations)

	sort.Slice(sortableConvs, func(i, j int) bool {
		return sortableConvs[i].LastActivityAt > sortableConvs[j].LastActivityAt
	})

	if limit > 0 && limit < len(sortableConvs) {
		sortableConvs = sortableConvs[:limit]
	}

	for _, conv := range sortableConvs {
		// Generate or retrieve internal ID
		internalID, exists := dm.internalIDMap[conv.ID]
		if !exists {
			internalID = fmt.Sprintf("%06d", dm.nextInternalID)
			dm.internalIDMap[conv.ID] = internalID
			dm.nextInternalID++
		}

		chat := &Chat{
			ID:           conv.ID,
			InternalID:   internalID,
			Title:        conv.Title,
			Users:        conv.Users,
			IsGroup:      conv.IsGroup,
			LastActivity: time.Unix(conv.LastActivityAt, 0),
		}

		// Get last message if available
		if len(conv.Items) > 0 {
			lastItem := conv.Items[0]
			if lastItem.Text != "" {
				chat.LastMessage = lastItem.Text
			}
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

// GetChatByInternalID finds a chat by its internal ID
func (dm *DirectMessages) GetChatByInternalID(internalID string) (*Chat, error) {
	chats, err := dm.GetChatsWithLimit(0)
	if err != nil {
		return nil, err
	}

	for _, chat := range chats {
		if chat.InternalID == internalID {
			return chat, nil
		}
	}

	return nil, fmt.Errorf("chat with internal ID %s not found", internalID)
}

// GetChatHistory fetches message history for a specific chat
func (dm *DirectMessages) GetChatHistory(chatID string, limit int) ([]*Message, error) {
	if dm.insta == nil {
		return nil, fmt.Errorf("not logged in")
	}

	chat, err := dm.GetChatByInternalID(chatID)
	if err == nil {
		chatID = chat.ID
	}

	var conversation *goinsta.Conversation
	for _, conv := range dm.insta.Inbox.Conversations {
		if conv.ID == chatID {
			conversation = conv
			break
		}
	}

	if conversation == nil {
		return nil, fmt.Errorf("chat not found")
	}

	// Get items if not already loaded
	if err := conversation.GetItems(); err != nil {
		return nil, fmt.Errorf("failed to get chat items: %v", err)
	}

	var messages []*Message
	itemCount := len(conversation.Items)
	if limit > 0 && limit < itemCount {
		itemCount = limit
	}

	for i := 0; i < itemCount; i++ {
		item := conversation.Items[i]
		message := &Message{
			ID:        item.ID,
			Text:      item.Text,
			Timestamp: time.Unix(item.Timestamp, 0),
			Type:      "text", // Default to text, could be enhanced
		}

		// Determine sender based on user ID comparison
		currentUserIDInt, _ := strconv.ParseInt(dm.currentUserID, 10, 64)
		if item.UserID == currentUserIDInt {
			message.Sender = "You"
		} else {
			// Try to find the user in the conversation's users list
			var senderName string
			for _, user := range conversation.Users {
				if user.ID == item.UserID {
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
			message.Sender = senderName
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// SendMessage sends a message to a specific chat
func (dm *DirectMessages) SendMessage(chatID, message string) error {
	if dm.insta == nil {
		return fmt.Errorf("not logged in")
	}

	chat, err := dm.GetChatByInternalID(chatID)
	if err == nil {
		chatID = chat.ID
	}

	// Find the conversation
	var conversation *goinsta.Conversation
	for _, conv := range dm.insta.Inbox.Conversations {
		if conv.ID == chatID {
			conversation = conv
			break
		}
	}

	if conversation == nil {
		return fmt.Errorf("chat not found")
	}

	// Send the message
	if err := conversation.Send(message); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// SendMessageToUser sends a message to a user by username
func (dm *DirectMessages) SendMessageToUser(username, message string) error {
	if dm.insta == nil {
		return fmt.Errorf("not logged in")
	}

	// Search for the user
	searchResult, err := dm.insta.Searchbar.SearchUser(username)
	if err != nil {
		return fmt.Errorf("failed to search for user: %v", err)
	}

	if len(searchResult.Users) == 0 {
		return fmt.Errorf("user not found: %s", username)
	}

	user := searchResult.Users[0]

	// Create a new conversation or find existing one
	_, err = dm.insta.Inbox.New(user, message)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// SendMessageByInternalID sends a message to a chat using its internal ID
func (dm *DirectMessages) SendMessageByInternalID(internalID, message string) error {
	chat, err := dm.GetChatByInternalID(internalID)
	if err != nil {
		return fmt.Errorf("chat with internal ID %s not found: %v", internalID, err)
	}

	return dm.SendMessage(chat.ID, message)
}

// SearchChats searches for chats by username or title
func (dm *DirectMessages) SearchChats(query string) ([]*Chat, error) {
	chats, err := dm.GetChats()
	if err != nil {
		return nil, err
	}

	var results []*Chat
	query = strings.ToLower(query)

	for _, chat := range chats {
		// Search by title
		if strings.Contains(strings.ToLower(chat.Title), query) {
			results = append(results, chat)
			continue
		}

		// Search by username
		for _, user := range chat.Users {
			if strings.Contains(strings.ToLower(user.Username), query) {
				results = append(results, chat)
				break
			}
		}
	}

	return results, nil
}

// MarkAsSeen marks a chat as seen
func (dm *DirectMessages) MarkAsSeen(chatID string) error {
	if dm.insta == nil {
		return fmt.Errorf("not logged in")
	}

	// Find the conversation
	var conversation *goinsta.Conversation
	for _, conv := range dm.insta.Inbox.Conversations {
		if conv.ID == chatID {
			conversation = conv
			break
		}
	}

	if conversation == nil {
		return fmt.Errorf("chat not found")
	}

	// Mark as seen - TODO: Fix this when we understand the MarkAsSeen API better
	// For now, we'll just return success
	_ = conversation

	return nil
}

// GetUnreadCount returns the total number of unread messages
func (dm *DirectMessages) GetUnreadCount() (int, error) {
	if dm.insta == nil {
		return 0, fmt.Errorf("not logged in")
	}

	if err := dm.insta.Inbox.Sync(); err != nil {
		return 0, fmt.Errorf("failed to sync inbox: %v", err)
	}

	return dm.insta.Inbox.UnseenCount, nil
}

// StartInteractiveChat starts an interactive chat session for a specific chat
func (dm *DirectMessages) StartInteractiveChat(chatID string) error {
	interactiveChat := NewInteractiveChat(dm, chatID)
	return interactiveChat.Start()
}
