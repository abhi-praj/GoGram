package chat

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StatusBar displays status messages and current mode
type StatusBar struct {
	*tview.TextView
	mode       ChatMode
	message    string
	defaultMsg string
	mutex      sync.RWMutex
	app        *tview.Application
}

// NewStatusBar creates a new status bar
func NewStatusBar(app *tview.Application) *StatusBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	sb := &StatusBar{
		TextView:   tv,
		app:        app,
		defaultMsg: "Ready",
	}

	// Set up the text view
	tv.SetBorder(false)
	tv.SetBackgroundColor(tcell.ColorDarkBlue)
	tv.SetTextColor(tcell.ColorWhite)

	return sb
}

// Update updates the status bar with a new message
func (sb *StatusBar) Update(message string) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()

	if message != "" {
		sb.message = message
	} else {
		sb.message = sb.defaultMsg
	}

	sb.app.QueueUpdateDraw(func() {
		sb.SetText(sb.message)
	})
}

// SetDefaultMessage sets the default message to show when no specific message is set
func (sb *StatusBar) SetDefaultMessage(msg string) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.defaultMsg = msg
}

// SetMode sets the chat mode and updates the status bar accordingly
func (sb *StatusBar) SetMode(mode ChatMode) {
	sb.mutex.Lock()
	defer sb.mutex.Unlock()
	sb.mode = mode

	var modeText string
	switch mode {
	case ChatModeChat:
		modeText = "CHAT MODE"
	case ChatModeCommand:
		modeText = "COMMAND MODE"
	case ChatModeReply:
		modeText = "REPLY MODE - Select message to reply to"
	case ChatModeUnsend:
		modeText = "UNSEND MODE - Select message to unsend"
	}

	sb.defaultMsg = modeText
	if sb.message == "" {
		sb.message = modeText
	}

	sb.app.QueueUpdateDraw(func() {
		sb.SetText(sb.message)
	})
}

// GetMode returns the current chat mode
func (sb *StatusBar) GetMode() ChatMode {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.mode
}

// GetMessage returns the current status message
func (sb *StatusBar) GetMessage() string {
	sb.mutex.RLock()
	defer sb.mutex.RUnlock()
	return sb.message
}
