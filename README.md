# GoGram - Instagram CLI Client

A Go implementation of Instagram CLI client, built with the goal of providing a terminal-based interface for Instagram messaging.

## Features

- **Authentication**: Login with username/password or saved sessions
- **Session Management**: Automatic session saving and loading
- **Configuration**: YAML-based configuration with dot notation support
- **Account Switching**: Switch between multiple Instagram accounts
- **2FA Support**: Built-in support for two-factor authentication
- **Interactive Chat**: Real-time messaging with automatic message updates

## Installation

### Prerequisites

- Go 1.21 or later
- Instagram account

### Build from Source

```bash
git clone <repository-url>
go mod tidy
go build -o ig-cli cmd/ig-cli/main.go
```

## Usage

### Basic Commands

```bash
# Display help and title
./ig-cli

# Show version
./ig-cli version

# Get help for auth commands
./ig-cli auth --help
```

### Authentication

```bash
# Login to Instagram
./ig-cli auth login

# Logout from Instagram
./ig-cli auth logout

# Switch to a different account
./ig-cli auth switch username

### Interactive Chat

```bash
# Open interactive chat with a specific chat ID
./ig-cli chat <chat_id>

# List available chats to find chat IDs
./ig-cli chat list
```
```

### Configuration

```bash
# List all configuration values
./ig-cli config list

# Get a specific configuration value
./ig-cli config get login.current_username

# Set a configuration value
./ig-cli config set login.current_username myusername

# Reset configuration to defaults
./ig-cli config reset
```

## Configuration

The application creates a configuration file at `~/.instagram-cli/config.yaml` with the following structure:

```yaml
language: en
login:
  default_username: null
  current_username: null
chat:
  layout: compact
  colors: true
scheduling:
  default_schedule_duration: "01:00"
privacy:
  invisible_mode: false
advanced:
  debug_mode: false
  data_dir: ~/.instagram-cli
  users_dir: ~/.instagram-cli/users
  cache_dir: ~/.instagram-cli/cache
  media_dir: ~/.instagram-cli/media
  generated_dir: ~/.instagram-cli/generated
```

## Session Management

Sessions are automatically saved to `~/.instagram-cli/users/<username>/session.json` after successful login. This allows you to stay logged in between application restarts.

## Interactive Chat

For detailed information about the interactive chat feature, see [INTERACTIVE_CHAT.md](INTERACTIVE_CHAT.md).

The interactive chat provides:
- Real-time messaging with automatic updates
- Last 10 messages displayed on entry
- Built-in commands for chat management
- Support for both direct messages and group chats

## Development

### Project Structure

```
ig-cli-go/
├── cmd/ig-cli/          # Main CLI entry point
├── internal/            # Internal packages
│   ├── auth/           # Authentication logic
│   ├── client/         # Instagram client wrapper
│   └── config/         # Configuration management
├── pkg/                # Public packages
└── go.mod              # Go module file
```

### Building

```bash
# Build for current platform
go build -o ig-cli cmd/ig-cli/main.go

# Build for specific platforms
GOOS=windows GOARCH=amd64 go build -o ig-cli.exe cmd/ig-cli/main.go
GOOS=linux GOARCH=amd64 go build -o ig-cli cmd/ig-cli/main.go
GOOS=darwin GOARCH=amd64 go build -o ig-cli cmd/ig-cli/main.go
```

## Dependencies

- [goinsta](https://github.com/Davincible/goinsta) - Instagram API client
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [YAML](https://gopkg.in/yaml.v3) - YAML parsing

## License

This project is licensed under the MIT License.

## Disclaimer

This project is not affiliated with, authorized, or endorsed by Instagram or any of its affiliates or subsidiaries. This is an independent and unofficial project. Use at your own risk.
