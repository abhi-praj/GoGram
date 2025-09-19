package chat

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
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
	dmInstance           *DirectMessages
}

// NewChatInterface creates a new chat interface
func NewChatInterface(app *tview.Application, onMessageSend func(string, string) error, onReplySend func(string, string, string) error, onUnsendMessage func(string) error, dmInstance *DirectMessages) *ChatInterface {
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
		dmInstance:           dmInstance,
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
	// Create main horizontal layout
	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Create left panel for chat list
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(ci.chatMenu, 0, 1, true)
	leftPanel.AddItem(ci.chatMenu.GetSearchInput(), 3, 0, false)

	// Create right panel for chat and input
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(ci.chatWindow, 0, 1, false)
	rightPanel.AddItem(ci.inputBox, 3, 0, false)
	rightPanel.AddItem(ci.statusBar, 1, 0, false)

	// Add panels to main layout (30% left, 70% right)
	mainFlex.AddItem(leftPanel, 0, 3, true)
	mainFlex.AddItem(rightPanel, 0, 7, false)

	// Set the root
	ci.app.SetRoot(mainFlex, true)

	// Set up global key handlers
	ci.app.SetInputCapture(ci.handleGlobalKeys)

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
	if chat == nil {
		ci.statusBar.Update("Error: No chat selected")
		return
	}

	ci.statusBar.Update(fmt.Sprintf("Loading chat: %s...", chat.Title))
	ci.SetCurrentChat(chat)
	ci.loadChatMessages(chat)
	ci.app.SetFocus(ci.inputBox)
}

// loadChatMessages loads messages for the selected chat
func (ci *ChatInterface) loadChatMessages(chat *Chat) {
	if ci.dmInstance == nil {
		ci.statusBar.Update("Error: DM instance not available")
		return
	}

	ci.statusBar.Update("Loading messages...")

	// Load messages in a goroutine to avoid blocking the UI
	go func() {
		messages, err := ci.dmInstance.GetChatHistory(chat.InternalID, ci.messagesPerFetch)
		if err != nil {
			ci.app.QueueUpdateDraw(func() {
				ci.statusBar.Update(fmt.Sprintf("Failed to load messages: %v", err))
			})
			return
		}

		// Update UI on main thread
		ci.app.QueueUpdateDraw(func() {
			ci.SetMessages(messages)
			ci.statusBar.Update(fmt.Sprintf("Loaded %d messages for %s", len(messages), chat.Title))
		})
	}()
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
				// Refresh messages after sending
				go func() {
					time.Sleep(500 * time.Millisecond) // Small delay to allow message to be processed
					ci.loadChatMessages(ci.currentChat)
				}()
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
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ci.stopRefresh:
			return
		case <-ticker.C:
			if ci.refreshEnabled && ci.currentChat != nil && ci.dmInstance != nil {
				ci.refreshLock.Lock()
				go func() {
					defer ci.refreshLock.Unlock()

					// Fetch latest messages
					messages, err := ci.dmInstance.GetChatHistory(ci.currentChat.InternalID, ci.messagesPerFetch)
					if err != nil {
						return
					}

					// Update UI on main thread
					ci.app.QueueUpdateDraw(func() {
						ci.SetMessages(messages)
					})
				}()
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

// handleGlobalKeys handles global keyboard shortcuts
func (ci *ChatInterface) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlQ:
		// Quit application
		ci.app.Stop()
		return nil
	case tcell.KeyTab:
		// Switch focus between panels
		if ci.app.GetFocus() == ci.chatMenu {
			ci.app.SetFocus(ci.inputBox)
		} else {
			ci.app.SetFocus(ci.chatMenu)
		}
		return nil
	case tcell.KeyCtrlR:
		// Refresh current chat
		if ci.currentChat != nil {
			ci.loadChatMessages(ci.currentChat)
		}
		return nil
	}
	return event
}

// Run starts the chat interface
func (ci *ChatInterface) Run() error {
	// Set initial status message after app starts
	go func() {
		ci.statusBar.Update("Welcome to IG-TUI! Use Tab to switch panels, Ctrl+Q to quit, Ctrl+R to refresh")
	}()

	ci.StartRefresh()
	return ci.app.Run()
}
