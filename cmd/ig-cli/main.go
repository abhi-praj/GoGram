package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/abhi-praj/GoGram/internal/auth"
	"github.com/abhi-praj/GoGram/internal/chat"
	"github.com/abhi-praj/GoGram/internal/client"
	"github.com/abhi-praj/GoGram/internal/config"
)

var (
	version        = "0.1.0"
	authInstance   *auth.InstagramAuth
	clientInstance *client.ClientWrapper
	dmInstance     *chat.DirectMessages
)

func main() {
	displayTitle()

	// Initialize auth
	authInstance = auth.NewInstagramAuth()

	// Start interactive shell
	startShell()
}

func displayTitle() {
	fmt.Print(`
   ██████╗  ██████╗   ██████╗ ██████╗  █████╗ ███╗   ███╗
  ██╔════╝ ██╔═══██╗ ██╔════╝ ██╔══██╗██╔══██╗████╗ ████║
  ██║  ███╗██║   ██║ ██║  ███╗██████╔╝███████║██╔████╔██║
  ██║   ██║██║   ██║ ██║   ██║██╔══██╗██╔══██║██║╚██╔╝██║
  ╚██████╔╝╚██████╔╝ ╚██████╔╝██║  ██║██║  ██║██║ ╚═╝ ██║
   ╚═════╝  ╚═════╝   ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝
`)
	fmt.Println("For the love of the game")
	fmt.Println()
	fmt.Println("Type 'help' to see available commands.")
	fmt.Printf("Version: %s\n\n", version)
}

func startShell() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("ig-cli> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\nReceived EOF. This usually means stdin was closed.")
				fmt.Println("Attempting to recover...")

				// Try to recreate the reader
				reader = bufio.NewReader(os.Stdin)
				fmt.Println("Reader recreated. Please try your command again.")
				continue
			}
			fmt.Printf("Error reading input: %v\n", err)
			fmt.Println("Continuing... Press Enter to continue or Ctrl+C to exit.")
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Check for exit command first
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Parse and execute command
		if err := executeCommand(input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func executeCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "help":
		showHelp()
	case "version":
		fmt.Printf("InstagramCLI v%s\n", version)
	case "login":
		return handleLogin()
	case "logout":
		return handleLogout()
	case "status":
		showStatus()
	case "chat":
		return handleChatCommand(args)
	case "config":
		return handleConfigCommand(args)
	case "clear":
		clearScreen()
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
	}

	return nil
}

func showHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                    - Show this help message")
	fmt.Println("  version                 - Show version information")
	fmt.Println("  login                   - Login to Instagram")
	fmt.Println("  logout                  - Logout from Instagram")
	fmt.Println("  status                  - Show current login status")
	fmt.Println("  chat <id>               - Open interactive chat with chat ID")
	fmt.Println("  chat list               - List recent chats (last 5)")
	fmt.Println("  chat list all           - List all chats")
	fmt.Println("  chat send <id> <msg>    - Send message to chat by ID")
	fmt.Println("  chat history <id>       - Show chat history")
	fmt.Println("  chat search <query>     - Search chats")
	fmt.Println("  config list             - List configuration values")
	fmt.Println("  config get <key>        - Get configuration value")
	fmt.Println("  config set <key> <val>  - Set configuration value")
	fmt.Println("  clear                   - Clear screen")
	fmt.Println("  exit/quit               - Exit the application")
	fmt.Println()
}

func handleLogin() error {
	fmt.Println("Attempting to login...")

	client, err := authInstance.Login()
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	clientInstance = client
	dmInstance = chat.NewDirectMessages(client)

	return nil
}

func handleLogout() error {
	if clientInstance == nil {
		return fmt.Errorf("not logged in")
	}

	if err := authInstance.Logout(""); err != nil {
		return fmt.Errorf("logout failed: %v", err)
	}

	clientInstance = nil
	dmInstance = nil
	return nil
}

func showStatus() {
	if clientInstance == nil {
		fmt.Println("Status: Not logged in")
		return
	}

	fmt.Printf("Status: Logged in as @%s\n", clientInstance.GetUsername())

	// Show unread count if available
	if dmInstance != nil {
		if count, err := dmInstance.GetUnreadCount(); err == nil {
			fmt.Printf("Unread messages: %d\n", count)
		}
	}
}

func handleChatCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: chat <command> [args]")
		fmt.Println("Commands: <id>, list, send, history, search")
		fmt.Println("  <id> - Open interactive chat with chat ID")
		return nil
	}

	if clientInstance == nil {
		return fmt.Errorf("not logged in. Use 'login' first.")
	}

	if !chat.IsSubcommand(args[0]) {
		// just make it an interactive chat if theres an id and nothing else
		return startInteractiveChat(args[0])
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "list":
		if len(args) > 1 && args[1] == "all" {
			return listAllChats()
		}
		return listChats()
	case "send":
		if len(args) < 3 {
			return fmt.Errorf("usage: chat send <chat_id> <message>")
		}
		return sendMessage(args[1], strings.Join(args[2:], " "))
	case "history":
		if len(args) < 2 {
			return fmt.Errorf("usage: chat history <chat_id>")
		}
		return showChatHistory(args[1])
	case "search":
		if len(args) < 2 {
			return fmt.Errorf("usage: chat search <query>")
		}
		return searchChats(args[1])
	default:
		fmt.Printf("Unknown chat command: %s\n", subcommand)
		fmt.Println("Available commands: <id>, list, send, history, search")
	}

	return nil
}

func listChats() error {
	chats, err := dmInstance.GetChats()
	if err != nil {
		return fmt.Errorf("failed to get chats: %v", err)
	}

	if len(chats) == 0 {
		fmt.Println("No chats found.")
		return nil
	}

	fmt.Printf("Found %d chats:\n", len(chats))
	fmt.Printf("%-8s %-20s %s\n", "ID", "Title", "Last Message")
	fmt.Printf("%-8s %-20s %s\n", "--", "-----", "------------")

	for _, chat := range chats {
		lastMsg := chat.LastMessage
		if lastMsg == "" {
			lastMsg = "(no message)"
		} else if len(lastMsg) > 30 {
			lastMsg = lastMsg[:27] + "..."
		}

		// Truncate title if too long
		title := chat.Title
		if len(title) > 18 {
			title = title[:15] + "..."
		}

		fmt.Printf("%-8s %-20s %s\n", chat.InternalID, title, lastMsg)
	}

	return nil
}

func sendMessage(chatID, message string) error {
	if err := dmInstance.SendMessageByInternalID(chatID, message); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	fmt.Printf("Message sent to chat %s\n", chatID)
	return nil
}

func showChatHistory(chatID string) error {
	messages, err := dmInstance.GetChatHistory(chatID, 20)
	if err != nil {
		return fmt.Errorf("failed to get chat history: %v", err)
	}

	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	fmt.Printf("Chat history (showing last %d messages):\n", len(messages))
	fmt.Printf("%-20s %-15s %s\n", "Time", "Sender", "Message")
	fmt.Printf("%-20s %-15s %s\n", "----", "------", "-------")

	for _, msg := range messages {
		timeStr := msg.Timestamp.Format("2006-01-02 15:04:05")
		fmt.Printf("%-20s %-15s %s\n", timeStr, msg.Sender, msg.Text)
	}

	return nil
}

func searchChats(query string) error {
	chats, err := dmInstance.SearchChats(query)
	if err != nil {
		return fmt.Errorf("failed to search chats: %v", err)
	}

	if len(chats) == 0 {
		fmt.Printf("No chats found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d chats matching '%s':\n", len(chats), query)
	fmt.Printf("%-8s %-20s %s\n", "ID", "Title", "Last Message")
	fmt.Printf("%-8s %-20s %s\n", "--", "-----", "------------")

	for _, chat := range chats {
		lastMsg := chat.LastMessage
		if len(lastMsg) > 30 {
			lastMsg = lastMsg[:27] + "..."
		}
		fmt.Printf("%-8s %-20s %s\n", chat.InternalID, chat.Title, lastMsg)
	}

	return nil
}

func listAllChats() error {
	chats, err := dmInstance.GetChatsWithLimit(0) // 0 means no limit
	if err != nil {
		return fmt.Errorf("failed to get chats: %v", err)
	}

	if len(chats) == 0 {
		fmt.Println("No chats found.")
		return nil
	}

	fmt.Printf("Found %d chats:\n", len(chats))
	fmt.Printf("%-8s %-20s %s\n", "ID", "Title", "Last Message")
	fmt.Printf("%-8s %-20s %s\n", "--", "-----", "------------")

	for _, chat := range chats {
		lastMsg := chat.LastMessage
		if lastMsg == "" {
			lastMsg = "(no message)"
		} else if len(lastMsg) > 30 {
			lastMsg = lastMsg[:27] + "..."
		}

		// Truncate title if too long
		title := chat.Title
		if len(title) > 18 {
			title = title[:15] + "..."
		}

		fmt.Printf("%-8s %-20s %s\n", chat.InternalID, title, lastMsg)
	}

	return nil
}

func handleConfigCommand(args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: config <command> [args]")
		fmt.Println("Commands: list, get, set")
		return nil
	}

	subcommand := strings.ToLower(args[0])
	cfg := config.GetInstance()

	switch subcommand {
	case "list":
		values := cfg.List()
		for _, kv := range values {
			fmt.Printf("%s = %v\n", kv.Key, kv.Value)
		}
	case "get":
		if len(args) < 2 {
			return fmt.Errorf("usage: config get <key>")
		}
		value := cfg.Get(args[1], nil)
		if value != nil {
			fmt.Println(value)
		} else {
			fmt.Printf("Configuration key '%s' not found\n", args[1])
		}
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: config set <key> <value>")
		}
		if err := cfg.Set(args[1], args[2]); err != nil {
			return fmt.Errorf("failed to set config: %v", err)
		}
		fmt.Printf("✅ Set %s = %s\n", args[1], args[2])
	default:
		fmt.Printf("Unknown config command: %s\n", subcommand)
		fmt.Println("Available commands: list, get, set")
	}

	return nil
}

func clearScreen() {
	// Simple clear for Windows
	fmt.Print("\033[H\033[2J")
}

// startInteractiveChat starts an interactive chat session
func startInteractiveChat(chatID string) error {
	fmt.Printf("Starting interactive chat with ID: %s\n", chatID)
	fmt.Println("Loading chat...")

	if err := dmInstance.StartInteractiveChat(chatID); err != nil {
		return fmt.Errorf("failed to start interactive chat: %v", err)
	}

	return nil
}
