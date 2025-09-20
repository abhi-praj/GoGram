package chat

import (
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputBox handles multi-line text input for chat messages
type InputBox struct {
	*tview.InputField
	buffer        []rune
	cursorPos     int
	scrollOffset  int
	currentHeight int
	lastHeight    int
	placeholder   string
	maxHeight     int
	mutex         sync.RWMutex
	app           *tview.Application
	onSubmit      func(string)
}

// NewInputBox creates a new input box
func NewInputBox(app *tview.Application, onSubmit func(string)) *InputBox {
	input := tview.NewInputField().
		SetLabel("Message: ").
		SetPlaceholder("Type a message...").
		SetFieldWidth(0)

	ib := &InputBox{
		InputField:  input,
		buffer:      make([]rune, 0),
		placeholder: "Type a message...",
		maxHeight:   5,
		app:         app,
		onSubmit:    onSubmit,
	}

	// Set up the input field
	input.SetBorder(true)
	input.SetTitle("Input")
	input.SetTitleAlign(tview.AlignCenter)

	// Set up input handling
	input.SetDoneFunc(ib.handleDone)

	return ib
}

// handleDone processes when the input is done (Enter pressed)
func (ib *InputBox) handleDone(key tcell.Key) {
	ib.mutex.Lock()
	defer ib.mutex.Unlock()

	switch key {
	case tcell.KeyEnter:
		ib.submitMessage()
	case tcell.KeyEscape:
		ib.clear()
	}
}

// insertRune inserts a rune at the current cursor position
func (ib *InputBox) insertRune(r rune) {
	if ib.cursorPos == len(ib.buffer) {
		ib.buffer = append(ib.buffer, r)
	} else {
		ib.buffer = append(ib.buffer[:ib.cursorPos], append([]rune{r}, ib.buffer[ib.cursorPos:]...)...)
	}
	ib.cursorPos++
	ib.updateDisplay()
}

// insertNewline inserts a newline character
func (ib *InputBox) insertNewline() {
	ib.insertRune('\n')
}

// handleBackspace handles backspace key
func (ib *InputBox) handleBackspace() {
	if ib.cursorPos > 0 {
		ib.buffer = append(ib.buffer[:ib.cursorPos-1], ib.buffer[ib.cursorPos:]...)
		ib.cursorPos--
		ib.updateDisplay()
	}
}

// handleDelete handles delete key
func (ib *InputBox) handleDelete() {
	if ib.cursorPos < len(ib.buffer) {
		ib.buffer = append(ib.buffer[:ib.cursorPos], ib.buffer[ib.cursorPos+1:]...)
		ib.updateDisplay()
	}
}

// moveCursorLeft moves cursor left
func (ib *InputBox) moveCursorLeft() {
	if ib.cursorPos > 0 {
		ib.cursorPos--
		ib.updateDisplay()
	}
}

// moveCursorRight moves cursor right
func (ib *InputBox) moveCursorRight() {
	if ib.cursorPos < len(ib.buffer) {
		ib.cursorPos++
		ib.updateDisplay()
	}
}

// moveCursorUp moves cursor up one line
func (ib *InputBox) moveCursorUp() {
	row, _ := ib.calculateCursorPosition()
	if row > 0 {
		targetPos := ib.getPositionFromRowCol(row-1, ib.getCursorColumn())
		if targetPos != -1 {
			ib.cursorPos = targetPos
			ib.updateDisplay()
		}
	}
}

// moveCursorDown moves cursor down one line
func (ib *InputBox) moveCursorDown() {
	row, _ := ib.calculateCursorPosition()
	targetPos := ib.getPositionFromRowCol(row+1, ib.getCursorColumn())
	if targetPos != -1 {
		ib.cursorPos = targetPos
		ib.updateDisplay()
	}
}

// moveCursorToStart moves cursor to start of current line
func (ib *InputBox) moveCursorToStart() {
	row, _ := ib.calculateCursorPosition()
	targetPos := ib.getPositionFromRowCol(row, 0)
	if targetPos != -1 {
		ib.cursorPos = targetPos
		ib.updateDisplay()
	}
}

// moveCursorToEnd moves cursor to end of current line
func (ib *InputBox) moveCursorToEnd() {
	row, _ := ib.calculateCursorPosition()
	nextRowStart := ib.getPositionFromRowCol(row+1, 0)
	if nextRowStart == -1 {
		ib.cursorPos = len(ib.buffer)
	} else {
		ib.cursorPos = nextRowStart - 1
	}
	ib.updateDisplay()
}

// calculateCursorPosition calculates the cursor's row and column position
func (ib *InputBox) calculateCursorPosition() (int, int) {
	textBeforeCursor := string(ib.buffer[:ib.cursorPos])
	lines := ib.wrapText(textBeforeCursor)

	if len(lines) == 0 {
		return 0, 0
	}

	row := len(lines) - 1
	col := len(lines[row])
	return row, col
}

// getCursorColumn returns the current cursor column
func (ib *InputBox) getCursorColumn() int {
	_, col := ib.calculateCursorPosition()
	return col
}

// wrapText wraps text into lines based on available width
func (ib *InputBox) wrapText(text string) []string {
	lines := strings.Split(text, "\n")
	wrappedLines := make([]string, 0)

	for _, line := range lines {
		if len(line) <= 60 { // Approximate width
			wrappedLines = append(wrappedLines, line)
		} else {
			// Simple word wrapping
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine)+len(word)+1 <= 60 {
					if currentLine != "" {
						currentLine += " " + word
					} else {
						currentLine = word
					}
				} else {
					if currentLine != "" {
						wrappedLines = append(wrappedLines, currentLine)
					}
					currentLine = word
				}
			}
			if currentLine != "" {
				wrappedLines = append(wrappedLines, currentLine)
			}
		}
	}

	return wrappedLines
}

// getPositionFromRowCol converts row and column position to buffer index
func (ib *InputBox) getPositionFromRowCol(row, col int) int {
	text := string(ib.buffer)
	lines := ib.wrapText(text)

	if row < 0 || row >= len(lines) {
		return -1
	}

	pos := 0
	for i := 0; i < row; i++ {
		pos += len(lines[i]) + 1 // +1 for newline
	}

	pos = min(pos+col, len(ib.buffer))
	return pos
}

// updateDisplay updates the input field display
func (ib *InputBox) updateDisplay() {
	text := string(ib.buffer)
	lines := ib.wrapText(text)

	// Calculate actual height needed
	ib.currentHeight = min(max(len(lines), 1), ib.maxHeight)

	// Update the input field text
	ib.app.QueueUpdateDraw(func() {
		ib.SetText(text)
	})
}

// submitMessage submits the current message
func (ib *InputBox) submitMessage() {
	text := strings.TrimSpace(string(ib.buffer))
	if text != "" && ib.onSubmit != nil {
		ib.onSubmit(text)
		ib.clear()
	}
}

// clear clears the input buffer
func (ib *InputBox) clear() {
	ib.mutex.Lock()
	defer ib.mutex.Unlock()

	ib.buffer = make([]rune, 0)
	ib.cursorPos = 0
	ib.scrollOffset = 0
	ib.currentHeight = 1
	ib.lastHeight = 1

	ib.app.QueueUpdateDraw(func() {
		ib.SetText("")
	})
}

// GetText returns the current text in the buffer
func (ib *InputBox) GetText() string {
	ib.mutex.RLock()
	defer ib.mutex.RUnlock()
	return string(ib.buffer)
}

// SetText sets the text in the buffer
func (ib *InputBox) SetText(text string) {
	ib.mutex.Lock()
	defer ib.mutex.Unlock()

	ib.buffer = []rune(text)
	ib.cursorPos = len(ib.buffer)
	ib.updateDisplay()
}
