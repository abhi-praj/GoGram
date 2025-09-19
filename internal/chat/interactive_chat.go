package chat

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Davincible/goinsta/v3"
)

// InteractiveChat handles real-time chat functionality
type InteractiveChat struct {
	dm           *DirectMessages
	chatID       string
	conversation *goinsta.Conversation
	reader       *bufio.Reader
	stopChan     chan bool
	mutex        sync.Mutex
	lastSentText string
}

// NewInteractiveChat creates a new interactive chat instance
func NewInteractiveChat(dm *DirectMessages, chatID string) *InteractiveChat {
	return &InteractiveChat{
		dm:           dm,
		chatID:       chatID,
		reader:       bufio.NewReader(os.Stdin),
		stopChan:     make(chan bool),
		lastSentText: "",
	}
}

// IsSubcommand checks if the given string is a known subcommand
func IsSubcommand(arg string) bool {
	subcommands := []string{"list"}
	arg = strings.ToLower(arg)
	for _, sub := range subcommands {
		if arg == sub {
			return true
		}
	}
	return false
}

// Start begins the interactive chat session
func (ic *InteractiveChat) Start() error {
	// Get chat details
	chat, err := ic.dm.GetChatByInternalID(ic.chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %v", err)
	}

	// Find the conversation
	var conversation *goinsta.Conversation
	for _, conv := range ic.dm.insta.Inbox.Conversations {
		if conv.ID == chat.ID {
			conversation = conv
			break
		}
	}

	if conversation == nil {
		return fmt.Errorf("conversation not found")
	}

	ic.conversation = conversation

	// Display chat header
	ic.displayChatHeader(chat)

	// Show last 10 messages
	if err := ic.displayRecentMessages(10); err != nil {
		fmt.Printf("Warning: Could not load recent messages: %v\n", err)
	}

	fmt.Println("\nChat started! Type your message and press Enter.")
	fmt.Println("Commands: /quit to exit, /help for help")
	fmt.Println("─" + strings.Repeat("─", 50))

	// Start message receiver in background
	go ic.messageReceiver()

	// Start input handler
	return ic.inputHandler()
}

// displayChatHeader shows the chat information header
func (ic *InteractiveChat) displayChatHeader(chat *Chat) {
	fmt.Printf("\nChat: %s\n", chat.Title)
	if chat.IsGroup {
		fmt.Printf("Group chat (%d members)\n", len(chat.Users))
	} else {
		fmt.Printf("Direct message with %s\n", chat.Users[0].Username)
	}
	fmt.Printf("Internal ID: %s\n", chat.InternalID)
	fmt.Println("─" + strings.Repeat("─", 50))
}

// displayRecentMessages shows the last N messages in the chat
func (ic *InteractiveChat) displayRecentMessages(limit int) error {
	messages, err := ic.dm.GetChatHistory(ic.chatID, limit)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		fmt.Println("No previous messages in this chat.")
		return nil
	}

	fmt.Printf("\nRecent messages:\n")
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		ic.displayMessage(msg, false)
	}

	return nil
}

// displayMessage displays a single message with proper formatting
func (ic *InteractiveChat) displayMessage(msg *Message, isNew bool) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	timeStr := msg.Timestamp.Format("15:04")

	if isNew {
		fmt.Printf("\n🆕 ")
	} else {
		fmt.Printf("\n")
	}

	if msg.Sender == "You" {
		fmt.Printf("You (%s): %s\n", timeStr, msg.Text)
	} else {
		fmt.Printf("%s (%s): %s\n", msg.Sender, timeStr, msg.Text)
	}
}

// inputHandler processes user input for sending messages
func (ic *InteractiveChat) inputHandler() error {
	for {
		select {
		case <-ic.stopChan:
			return nil
		default:
			fmt.Print("You: ")
			input, err := ic.reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input: %v", err)
			}

			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			// Handle commands
			if strings.HasPrefix(input, "/") {
				if err := ic.handleCommand(input); err != nil {
					fmt.Printf("Command error: %v\n", err)
				}
				continue
			}

			// Send message
			if err := ic.sendMessage(input); err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
			} else {
				// Track the sent message to avoid duplicate display
				ic.lastSentText = input
				// Show subtle sending indicator that will be replaced by the actual message
				fmt.Printf("Sending...\n")
			}
		}
	}
}

// handleCommand processes chat commands
func (ic *InteractiveChat) handleCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "/quit", "/exit":
		fmt.Println("Exiting chat...")
		close(ic.stopChan)
		return nil
	case "/help":
		ic.showHelp()
	case "/clear":
		ic.clearScreen()
	case "/refresh":
		if err := ic.displayRecentMessages(10); err != nil {
			fmt.Printf("Failed to refresh: %v\n", err)
		} else {
			fmt.Println("Chat refreshed")
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		ic.showHelp()
	}

	return nil
}

// showHelp displays available commands
func (ic *InteractiveChat) showHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  /quit, /exit  - Exit the chat")
	fmt.Println("  /help         - Show this help")
	fmt.Println("  /clear        - Clear the screen")
	fmt.Println("  /refresh      - Refresh recent messages")
	fmt.Println("  (type message) - Send a message")
}

// clearScreen clears the terminal
func (ic *InteractiveChat) clearScreen() {
	fmt.Print("\033[H\033[2J")
	ic.displayChatHeader(&Chat{Title: "Chat", InternalID: ic.chatID})
}

// sendMessage sends a message to the current chat
func (ic *InteractiveChat) sendMessage(text string) error {
	return ic.dm.SendMessageByInternalID(ic.chatID, text)
}

// messageReceiver continuously checks for new messages
func (ic *InteractiveChat) messageReceiver() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ic.stopChan:
			return
		case <-ticker.C:
			ic.checkForNewMessages()
		}
	}
}

// checkForNewMessages checks if there are new messages and displays them
func (ic *InteractiveChat) checkForNewMessages() {
	// Sync inbox to get latest messages
	if err := ic.dm.insta.Inbox.Sync(); err != nil {
		return // Silently fail, will retry next tick
	}

	// Find our conversation
	var conversation *goinsta.Conversation
	for _, conv := range ic.dm.insta.Inbox.Conversations {
		if conv.ID == ic.conversation.ID {
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

	// Check for new messages (messages newer than our last check)
	if len(conversation.Items) > 0 {
		latestItem := conversation.Items[0]
		latestTime := time.Unix(latestItem.Timestamp, 0)

		currentUserIDInt, _ := strconv.ParseInt(ic.dm.currentUserID, 10, 64)

		if latestTime.After(time.Now().Add(-10*time.Second)) &&
			latestItem.UserID != currentUserIDInt {

			msg := &Message{
				ID:        latestItem.ID,
				Text:      latestItem.Text,
				Timestamp: latestTime,
				Type:      "text",
			}

			// todo make sure the sender logic is correct
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

			ic.displayMessage(msg, true)
		}

		// Handle sent messages (messages from current user that were just sent)
		if latestItem.UserID == currentUserIDInt &&
			latestTime.After(time.Now().Add(-10*time.Second)) &&
			latestItem.Text == ic.lastSentText &&
			ic.lastSentText != "" {

			msg := &Message{
				ID:        latestItem.ID,
				Text:      latestItem.Text,
				Timestamp: latestTime,
				Type:      "text",
				Sender:    "You",
			}

			ic.displayMessage(msg, false)
			ic.lastSentText = ""
		}
	}
}
