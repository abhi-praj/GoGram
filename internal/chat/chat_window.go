package chat

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rivo/tview"
)

// ChatWindow handles chat message display and formatting
type ChatWindow struct {
	*tview.TextView
	messages             []*Message
	messagesLines        []*LineInfo
	selection            int
	selectedMessageID    string
	scrollOffset         int
	visibleMessagesRange [2]int
	visibleLinesRange    [2]int
	mode                 ChatMode
	mutex                sync.RWMutex
	app                  *tview.Application
}

// NewChatWindow creates a new chat window
func NewChatWindow(app *tview.Application) *ChatWindow {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	cw := &ChatWindow{
		TextView: tv,
		app:      app,
	}

	// Set up the text view
	tv.SetBorder(true)
	tv.SetTitle("Chat")
	tv.SetTitleAlign(tview.AlignCenter)

	return cw
}

// SetMessages updates the messages list
func (cw *ChatWindow) SetMessages(messages []*Message) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()

	cw.messages = messages
	cw.buildMessageLines()
}

// buildMessageLines builds wrapped lines for chat messages with word wrapping and formatting
func (cw *ChatWindow) buildMessageLines() {
	linesBuffer := make([]*LineInfo, 0)
	width := 80 // Default width, will be updated when drawing

	// Build wrapped lines from oldest to newest
	for msgIdx, msg := range cw.messages {
		senderText := msg.Sender + ": "
		senderWidth := len(senderText)

		// Handle the main message
		contentWidth := width - senderWidth - 1
		isSelected := msgIdx == cw.selection && cw.mode == ChatModeReply

		// Determine color index
		colorIdx := (hashString(msg.Sender) % 3) + 1

		// Split content into words, then chunk
		words := strings.Fields(msg.Text)
		lineBuffer := make([]string, 0)
		currentWidth := 0
		firstLine := true

		flushLine := func() {
			if len(lineBuffer) > 0 {
				lineText := strings.Join(lineBuffer, " ")
				if firstLine {
					linesBuffer = append(linesBuffer, &LineInfo{
						MessageIdx:  msgIdx,
						Text:        lineText,
						IsSelected:  isSelected,
						ColorIdx:    colorIdx,
						SenderWidth: senderWidth,
						SenderText:  senderText,
						IsDimmed:    false,
					})
				} else {
					linesBuffer = append(linesBuffer, &LineInfo{
						MessageIdx:  msgIdx,
						Text:        lineText,
						IsSelected:  isSelected,
						ColorIdx:    colorIdx,
						SenderWidth: senderWidth,
						SenderText:  " " + strings.Repeat(" ", senderWidth-1),
						IsDimmed:    false,
					})
				}
			}
		}

		for _, word := range words {
			spaceNeeded := 1
			if len(lineBuffer) == 0 {
				spaceNeeded = 0
			}
			if currentWidth+len(word)+spaceNeeded <= contentWidth {
				lineBuffer = append(lineBuffer, word)
				currentWidth += len(word) + spaceNeeded
			} else {
				flushLine()
				lineBuffer = []string{word}
				currentWidth = len(word)
				firstLine = false
			}
		}

		// Flush remaining line buffer
		flushLine()

		// Add a blank line after each message
		linesBuffer = append(linesBuffer, &LineInfo{
			MessageIdx:  msgIdx,
			Text:        "",
			IsSelected:  false,
			ColorIdx:    0,
			SenderWidth: 0,
			SenderText:  "",
			IsDimmed:    false,
		})
	}

	cw.messagesLines = linesBuffer
}

// Update refreshes the chat window display
func (cw *ChatWindow) Update() {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()

	if len(cw.messages) == 0 {
		return
	}

	// Update visible messages range
	if len(cw.messagesLines) > 0 {
		cw.visibleLinesRange = [2]int{
			max(0, len(cw.messagesLines)-cw.getHeight()-cw.scrollOffset),
			max(0, len(cw.messagesLines)-1-cw.scrollOffset),
		}
		if cw.visibleLinesRange[0] < len(cw.messagesLines) && cw.visibleLinesRange[1] < len(cw.messagesLines) {
			cw.visibleMessagesRange = [2]int{
				cw.messagesLines[cw.visibleLinesRange[0]].MessageIdx,
				cw.messagesLines[cw.visibleLinesRange[1]].MessageIdx,
			}
		}
	}

	// Build the display text
	var displayText strings.Builder

	// Print from the bottom up
	for i := len(cw.messagesLines) - 1; i >= 0; i-- {
		if i < cw.visibleLinesRange[0] || i > cw.visibleLinesRange[1] {
			continue
		}

		line := cw.messagesLines[i]

		// Add color formatting
		if line.IsSelected {
			displayText.WriteString("[red]")
		}
		if line.ColorIdx > 0 && !line.IsDimmed {
			displayText.WriteString(fmt.Sprintf("[%s]", getColorTag(line.ColorIdx)))
		}
		if line.IsDimmed {
			displayText.WriteString("[gray]")
		}

		// Add sender text
		if line.SenderText != "" {
			displayText.WriteString(line.SenderText)
		}

		// Add message text
		displayText.WriteString(line.Text)

		// Close color tags
		if line.IsSelected {
			displayText.WriteString("[-]")
		}
		if line.ColorIdx > 0 && !line.IsDimmed {
			displayText.WriteString("[-]")
		}
		if line.IsDimmed {
			displayText.WriteString("[-]")
		}

		displayText.WriteString("\n")
	}

	// Update the text view
	cw.app.QueueUpdateDraw(func() {
		cw.SetText(displayText.String())
		cw.ScrollToEnd()
	})
}

// SetMode sets the chat mode
func (cw *ChatWindow) SetMode(mode ChatMode) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.mode = mode
}

// SetSelection sets the current selection
func (cw *ChatWindow) SetSelection(selection int) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.selection = selection
}

// GetSelection returns the current selection
func (cw *ChatWindow) GetSelection() int {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()
	return cw.selection
}

// SetSelectedMessageID sets the selected message ID for replies
func (cw *ChatWindow) SetSelectedMessageID(id string) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.selectedMessageID = id
}

// GetSelectedMessageID returns the selected message ID
func (cw *ChatWindow) GetSelectedMessageID() string {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()
	return cw.selectedMessageID
}

// ScrollUp scrolls up in the chat
func (cw *ChatWindow) ScrollUp() {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.scrollOffset = min(cw.scrollOffset+cw.getHeight()-1, len(cw.messagesLines)-cw.getHeight())
}

// ScrollDown scrolls down in the chat
func (cw *ChatWindow) ScrollDown() {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.scrollOffset = max(cw.scrollOffset-cw.getHeight()+1, 0)
}

// getHeight returns the height of the text view
func (cw *ChatWindow) getHeight() int {
	_, _, _, height := cw.GetRect()
	return height - 2 // Account for border
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// hashString creates a simple hash for color assignment
func hashString(s string) int {
	hash := 0
	for _, char := range s {
		hash = (hash*31 + int(char)) % 3
	}
	return hash
}

// getColorTag returns the color tag for tview
func getColorTag(idx int) string {
	colors := []string{"red", "blue", "green"}
	if idx > 0 && idx <= len(colors) {
		return colors[idx-1]
	}
	return "white"
}
