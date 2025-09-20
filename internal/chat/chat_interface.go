package chat

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rivo/tview"
)

// ChatInterface is the main chat interface that coordinates components and handles user input
type ChatInterface struct {
	app                  *tview.Application
	chatWindow           *ChatWindow
	inputBox             *InputBox
	statusBar            *StatusBar
	chatMenu             *ChatMenu
	mode                 ChatMode
	height, width        int
	messagesPerFetch     int
	skipMessageSelection bool
	refreshLock          sync.Mutex
	mutex                sync.RWMutex
	stopRefresh          chan bool
	refreshEnabled       bool
	currentChat          *Chat
	onMessageSend        func(string, string) error
	onReplySend          func(string, string, string) error
	onUnsendMessage      func(string) error
}

// NewChatInterface creates a new chat interface
func NewChatInterface(app *tview.Application, onMessageSend func(string, string) error, onReplySend func(string, string, string) error, onUnsendMessage func(string) error) *ChatInterface {
	ci := &ChatInterface{
		app:                  app,
		mode:                 ChatModeChat,
		messagesPerFetch:     20,
		skipMessageSelection: false,
		stopRefresh:          make(chan bool),
		refreshEnabled:       true,
		onMessageSend:        onMessageSend,
		onReplySend:          onReplySend,
		onUnsendMessage:      onUnsendMessage,
	}

	// Initialize components
	ci.chatWindow = NewChatWindow(app)
	ci.inputBox = NewInputBox(app, ci.handleMessageSubmit)
	ci.statusBar = NewStatusBar(app)
	ci.chatMenu = NewChatMenu(app, ci.handleChatSelect)

	// Set up the layout
	ci.setupLayout()

	return ci
}

// setupLayout sets up the application layout
func (ci *ChatInterface) setupLayout() {
	// Create a flex layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add chat menu at the top (1/3 of height)
	flex.AddItem(ci.chatMenu, 0, 1, false)

	// Add search input below chat menu
	flex.AddItem(ci.chatMenu.GetSearchInput(), 3, 0, false)

	// Add chat window in the middle (2/3 of height)
	flex.AddItem(ci.chatWindow, 0, 2, false)

	// Add input box at the bottom
	flex.AddItem(ci.inputBox, 6, 0, false)

	// Add status bar at the very bottom
	flex.AddItem(ci.statusBar, 1, 0, false)

	// Set the root
	ci.app.SetRoot(flex, true)

	// Set focus to chat menu initially
	ci.app.SetFocus(ci.chatMenu)
}

// SetChats sets the chat list for the menu
func (ci *ChatInterface) SetChats(chats []*Chat) {
	ci.chatMenu.SetChats(chats)
}

// SetMessages sets the messages for the current chat
func (ci *ChatInterface) SetMessages(messages []*Message) {
	ci.chatWindow.SetMessages(messages)
	ci.chatWindow.Update()
}

// SetCurrentChat sets the current active chat
func (ci *ChatInterface) SetCurrentChat(chat *Chat) {
	ci.currentChat = chat
	ci.chatWindow.SetTitle(fmt.Sprintf("Chat: %s", chat.Title))
	ci.statusBar.Update(fmt.Sprintf("Active chat: %s", chat.Title))
}

// handleChatSelect handles when a chat is selected from the menu
func (ci *ChatInterface) handleChatSelect(chat *Chat) {
	ci.SetCurrentChat(chat)
	ci.app.SetFocus(ci.inputBox)
	ci.statusBar.Update(fmt.Sprintf("Switched to chat: %s", chat.Title))
}

// handleMessageSubmit handles message submission from the input box
func (ci *ChatInterface) handleMessageSubmit(message string) {
	if ci.currentChat == nil {
		ci.statusBar.Update("No chat selected")
		return
	}

	if strings.TrimSpace(message) == "" {
		return
	}

	// Check if we're in reply mode
	if ci.mode == ChatModeReply && ci.chatWindow.GetSelectedMessageID() != "" {
		// Send reply
		if ci.onReplySend != nil {
			ci.statusBar.Update("Sending reply...")
			if err := ci.onReplySend(ci.currentChat.InternalID, message, ci.chatWindow.GetSelectedMessageID()); err != nil {
				ci.statusBar.Update(fmt.Sprintf("Failed to send reply: %v", err))
			} else {
				ci.statusBar.Update("Reply sent")
				ci.chatWindow.SetSelectedMessageID("")
				ci.SetMode(ChatModeChat)
			}
		}
	} else {
		// Send regular message
		if ci.onMessageSend != nil {
			ci.statusBar.Update("Sending...")
			if err := ci.onMessageSend(ci.currentChat.InternalID, message); err != nil {
				ci.statusBar.Update(fmt.Sprintf("Failed to send message: %v", err))
			} else {
				ci.statusBar.Update("Message sent")
			}
		}
	}
}

// SetMode sets the chat mode
func (ci *ChatInterface) SetMode(mode ChatMode) {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()

	ci.mode = mode
	ci.chatWindow.SetMode(mode)
	ci.statusBar.SetMode(mode)
}

// StartRefresh starts the background refresh thread
func (ci *ChatInterface) StartRefresh() {
	go ci.refreshChat()
}

// StopRefresh stops the background refresh
func (ci *ChatInterface) StopRefresh() {
	close(ci.stopRefresh)
}

// refreshChat refreshes the chat messages in the background
func (ci *ChatInterface) refreshChat() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ci.stopRefresh:
			return
		case <-ticker.C:
			if ci.refreshEnabled && ci.currentChat != nil {
				ci.refreshLock.Lock()
				// Here you would typically fetch new messages
				// For now, we'll just update the display
				ci.chatWindow.Update()
				ci.refreshLock.Unlock()
			}
		}
	}
}

// ToggleRefresh enables/disables automatic message fetching
func (ci *ChatInterface) ToggleRefresh(enabled bool) {
	ci.refreshEnabled = enabled
}

// HandleCommand handles chat commands
func (ci *ChatInterface) HandleCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	switch cmd {
	case "reply":
		ci.SetMode(ChatModeReply)
		ci.statusBar.Update("Reply mode: Select a message to reply to")
	case "unsend":
		ci.SetMode(ChatModeUnsend)
		ci.statusBar.Update("Unsend mode: Select a message to unsend")
	case "chat":
		ci.SetMode(ChatModeChat)
		ci.statusBar.Update("Back to chat mode")
	case "help":
		ci.showHelp()
	default:
		ci.statusBar.Update(fmt.Sprintf("Unknown command: %s", cmd))
	}
}

// showHelp displays available commands
func (ci *ChatInterface) showHelp() {
	ci.statusBar.Update("Help displayed")
	// You could display this in a modal or in the chat window
}

// GetChatWindow returns the chat window component
func (ci *ChatInterface) GetChatWindow() *ChatWindow {
	return ci.chatWindow
}

// GetInputBox returns the input box component
func (ci *ChatInterface) GetInputBox() *InputBox {
	return ci.inputBox
}

// GetStatusBar returns the status bar component
func (ci *ChatInterface) GetStatusBar() *StatusBar {
	return ci.statusBar
}

// GetChatMenu returns the chat menu component
func (ci *ChatInterface) GetChatMenu() *ChatMenu {
	return ci.chatMenu
}

// Run starts the chat interface
func (ci *ChatInterface) Run() error {
	ci.StartRefresh()
	return ci.app.Run()
}
