# IG-TUI - Instagram Terminal User Interface

A Go implementation of Instagram Terminal User Interface (TUI), providing a modern terminal-based interface for Instagram messaging with real-time chat functionality.

## Features

- **Modern TUI Interface**: Beautiful terminal-based user interface with real-time updates
- **Authentication**: Login with username/password or saved sessions
- **Session Management**: Automatic session saving and loading
- **Real-time Messaging**: Send and receive messages with automatic refresh (every 3 seconds)
- **Split-panel Layout**: Chat list on the left, conversation on the right
- **Keyboard Navigation**: Tab to switch panels, Ctrl+Q to quit, Ctrl+R to refresh
- **Search Functionality**: Search chats by username (@) or title (/)
- **Configuration**: YAML-based configuration with dot notation support
- **Account Switching**: Switch between multiple Instagram accounts
- **2FA Support**: Built-in support for two-factor authentication

## Installation

### Prerequisites

- Go 1.21 or later
- Instagram account

### Build from Source

```bash
git clone https://github.com/abhi-praj/ig-tui
cd ig-tui
go build -o ig-tui cmd/ig-tui/main.go
```

## Usage

Simply run the TUI application:

```bash
./ig-tui
```

The application will:
1. Automatically handle login (prompts for credentials if needed)
2. Launch the TUI interface with your Instagram chats
3. Display a split-panel interface with chat list and conversation view

### TUI Interface

- **Left Panel**: Browse your Instagram conversations
- **Right Panel**: View messages and send new ones
- **Status Bar**: Shows helpful information and current status

### Keyboard Shortcuts

- **Tab**: Switch focus between chat list and input box
- **Ctrl+Q**: Quit the application
- **Ctrl+R**: Refresh current chat messages
- **Enter**: Select chat (in chat list) or send message (in input box)
- **Arrow Keys**: Navigate through chats
- **@username**: Search for chats by username
- **/title**: Search for chats by title

### How to Use

1. **Start the application**: Run `./ig-tui`
2. **Login**: Enter your Instagram credentials when prompted
3. **Browse chats**: Use arrow keys in the left panel to navigate
4. **Select a chat**: Press Enter to open a conversation
5. **Send messages**: Type in the input box and press Enter
6. **Switch panels**: Use Tab to move between chat list and input
7. **Search**: Type @ or / to search for specific chats

## Configuration

The application uses YAML-based configuration stored in `~/.config/ig-tui/config.yaml`.

### Configuration Options

```yaml
login:
  current_username: "your_username"
  save_session: true

chat:
  auto_refresh: true
  refresh_interval: 3s
  message_limit: 20

ui:
  theme: "default"
  show_timestamps: true
```

## Development

### Project Structure

```
ig-tui-go/
├── cmd/ig-tui/          # Main TUI entry point
├── internal/            # Internal packages
│   ├── auth/           # Authentication logic
│   ├── client/         # Instagram client wrapper
│   ├── chat/           # TUI chat interface
│   └── config/         # Configuration management
├── pkg/                # Public packages
└── go.mod              # Go module definition
```

### Building

```bash
# Build for current platform
go build -o ig-tui cmd/ig-tui/main.go

# Build for specific platforms
GOOS=windows GOARCH=amd64 go build -o ig-tui.exe cmd/ig-tui/main.go
GOOS=linux GOARCH=amd64 go build -o ig-tui cmd/ig-tui/main.go
GOOS=darwin GOARCH=amd64 go build -o ig-tui cmd/ig-tui/main.go
```

## Dependencies

- [goinsta](https://github.com/Davincible/goinsta) - Instagram API client
- [tview](https://github.com/rivo/tview) - Terminal UI framework
- [viper](https://github.com/spf13/viper) - Configuration management

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This tool is for educational purposes only. Use it responsibly and in accordance with Instagram's Terms of Service. The developers are not responsible for any misuse or violations of Instagram's policies.

## Troubleshooting

### Common Issues

1. **Login fails**: Make sure your Instagram credentials are correct and 2FA is properly handled
2. **Messages not loading**: Check your internet connection and Instagram account status
3. **TUI not displaying correctly**: Ensure your terminal supports the required features

### Support

If you encounter issues, please:
1. Check the troubleshooting section
2. Search existing issues on GitHub
3. Create a new issue with detailed information about your problem
