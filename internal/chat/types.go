package chat

// ChatMode represents the different modes of the chat interface
type ChatMode int

const (
	ChatModeChat ChatMode = iota
	ChatModeCommand
	ChatModeReply
	ChatModeUnsend
)

// Signal represents continue or quit chat
type Signal int

const (
	SignalContinue Signal = iota
	SignalBack
	SignalQuit
)

// ChatMenuMode represents different modes of the chat menu
type ChatMenuMode int

const (
	ChatMenuModeDefault ChatMenuMode = iota
	ChatMenuModeSearchUsername
	ChatMenuModeSearchTitle
)

// LineInfo stores line information for chat messages
type LineInfo struct {
	MessageIdx  int
	Text        string
	IsSelected  bool
	ColorIdx    int
	SenderWidth int
	SenderText  string
	IsDimmed    bool
}
