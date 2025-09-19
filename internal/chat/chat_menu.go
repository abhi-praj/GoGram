package chat

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ChatMenu displays the chat list and allows the user to select one
type ChatMenu struct {
	*tview.List
	chats        []*Chat
	selection    int
	scrollOffset int
	searchQuery  string
	placeholder  string
	mode         ChatMenuMode
	mutex        sync.RWMutex
	app          *tview.Application
	onChatSelect func(*Chat)
	searchInput  *tview.InputField
	statusBar    *tview.TextView
}

// NewChatMenu creates a new chat menu
func NewChatMenu(app *tview.Application, onChatSelect func(*Chat)) *ChatMenu {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetMainTextColor(tcell.ColorWhite).
		SetSelectedTextColor(tcell.ColorWhite).
		SetSelectedBackgroundColor(tcell.ColorBlue)

	cm := &ChatMenu{
		List:         list,
		chats:        make([]*Chat, 0),
		placeholder:  "Search for chat by @username or /title + ENTER",
		app:          app,
		onChatSelect: onChatSelect,
	}

	// Set up the list
	list.SetBorder(true)
	list.SetTitle("Chat List")
	list.SetTitleAlign(tview.AlignCenter)

	// Create search input
	cm.searchInput = tview.NewInputField()
	cm.searchInput.SetLabel("Search: ")
	cm.searchInput.SetPlaceholder(cm.placeholder)
	cm.searchInput.SetFieldWidth(0)
	cm.searchInput.SetBorder(true)
	cm.searchInput.SetTitle("Search")
	cm.searchInput.SetTitleAlign(tview.AlignCenter)

	// Create status bar
	cm.statusBar = tview.NewTextView()
	cm.statusBar.SetDynamicColors(true)
	cm.statusBar.SetTextAlign(tview.AlignCenter)
	cm.statusBar.SetBorder(false)
	cm.statusBar.SetBackgroundColor(tcell.ColorDarkBlue)
	cm.statusBar.SetTextColor(tcell.ColorWhite)

	// Set up input handling
	cm.searchInput.SetDoneFunc(cm.handleSearchDone)
	list.SetSelectedFunc(cm.handleChatSelect)

	// Add key handler for the list
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Stop()
			return nil
		case tcell.KeyTab:
			// Let the global handler deal with Tab
			return event
		case tcell.KeyCtrlR:
			// Let the global handler deal with Ctrl+R
			return event
		}
		return event
	})

	return cm
}

// SetChats updates the chat list
func (cm *ChatMenu) SetChats(chats []*Chat) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.chats = chats
	cm.updateChatList()
}

// updateChatList updates the displayed chat list
func (cm *ChatMenu) updateChatList() {
	cm.Clear()

	for _, chat := range cm.chats {
		title := chat.Title
		if chat.IsGroup {
			title = fmt.Sprintf("📱 %s (%d members)", title, len(chat.Users))
		} else {
			title = fmt.Sprintf("👤 %s", title)
		}

		// Add unread indicator
		if chat.UnreadCount > 0 {
			title = fmt.Sprintf("🔴 %s (%d unread)", title, chat.UnreadCount)
		}

		cm.AddItem(title, chat.InternalID, 0, nil)
	}

	// Set selection
	if len(cm.chats) > 0 {
		cm.SetCurrentItem(cm.selection)
	}
}

// handleSearchDone processes search input completion
func (cm *ChatMenu) handleSearchDone(key tcell.Key) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	switch key {
	case tcell.KeyEnter:
		query := strings.TrimSpace(cm.searchInput.GetText())
		if query != "" {
			cm.performSearch(query)
		}
		cm.searchInput.SetText("")
		cm.mode = ChatMenuModeDefault
		cm.updateStatusBar()
	case tcell.KeyEscape:
		cm.searchInput.SetText("")
		cm.mode = ChatMenuModeDefault
		cm.updateStatusBar()
	}
}

// performSearch performs the search operation
func (cm *ChatMenu) performSearch(query string) {
	// Simple search implementation - can be enhanced later
	var results []*Chat

	if strings.HasPrefix(query, "@") {
		// Search by username
		username := strings.TrimPrefix(query, "@")
		for _, chat := range cm.chats {
			for _, user := range chat.Users {
				if strings.Contains(strings.ToLower(user.Username), strings.ToLower(username)) {
					results = append(results, chat)
					break
				}
			}
		}
	} else if strings.HasPrefix(query, "/") {
		// Search by title
		title := strings.TrimPrefix(query, "/")
		for _, chat := range cm.chats {
			if strings.Contains(strings.ToLower(chat.Title), strings.ToLower(title)) {
				results = append(results, chat)
			}
		}
	} else {
		// General search
		for _, chat := range cm.chats {
			if strings.Contains(strings.ToLower(chat.Title), strings.ToLower(query)) {
				results = append(results, chat)
			} else {
				for _, user := range chat.Users {
					if strings.Contains(strings.ToLower(user.Username), strings.ToLower(query)) ||
						strings.Contains(strings.ToLower(user.FullName), strings.ToLower(query)) {
						results = append(results, chat)
						break
					}
				}
			}
		}
	}

	if len(results) > 0 {
		cm.chats = results
		cm.selection = 0
		cm.updateChatList()
		cm.updateStatusBar(fmt.Sprintf("Found %d results", len(results)))
	} else {
		cm.updateStatusBar("No results found")
	}
}

// handleChatSelect processes chat selection
func (cm *ChatMenu) handleChatSelect(index int, mainText, secondaryText string, shortcut rune) {
	if index >= 0 && index < len(cm.chats) {
		if cm.onChatSelect != nil {
			cm.onChatSelect(cm.chats[index])
		}
	}
}

// SetMode sets the chat menu mode
func (cm *ChatMenu) SetMode(mode ChatMenuMode) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.mode = mode
	cm.updateStatusBar()
}

// updateStatusBar updates the status bar message
func (cm *ChatMenu) updateStatusBar(message ...string) {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	} else {
		switch cm.mode {
		case ChatMenuModeDefault:
			msg = "Use arrow keys to navigate, Enter to select, @ to search username, / to search title"
		case ChatMenuModeSearchUsername:
			msg = "Type username and press Enter to search"
		case ChatMenuModeSearchTitle:
			msg = "Type title and press Enter to search"
		}
	}

	cm.app.QueueUpdateDraw(func() {
		cm.statusBar.SetText(msg)
	})
}

// GetSearchInput returns the search input field
func (cm *ChatMenu) GetSearchInput() *tview.InputField {
	return cm.searchInput
}

// GetStatusBar returns the status bar
func (cm *ChatMenu) GetStatusBar() *tview.TextView {
	return cm.statusBar
}

// SetSelection sets the current selection
func (cm *ChatMenu) SetSelection(selection int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.selection = selection
	if len(cm.chats) > 0 && cm.selection < len(cm.chats) {
		cm.SetCurrentItem(cm.selection)
	}
}

// GetSelection returns the current selection
func (cm *ChatMenu) GetSelection() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.selection
}

// GetChats returns the current chat list
func (cm *ChatMenu) GetChats() []*Chat {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.chats
}
