package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/abhi-praj/GoGram/internal/auth"
	"github.com/abhi-praj/GoGram/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	rootCmd = &cobra.Command{
		Use:   "ig-cli",
		Short: "Instagram CLI",
		Long: `Instagram CLI.

I don't have a corny bio for this project`,
		Run: func(cmd *cobra.Command, args []string) {
			startInteractiveMode()
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(configCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("InstagramCLI v%s\n", version)
	},
}

var interactiveCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start interactive shell mode",
	Long: `Start an interactive shell where you can type commands directly.
	
Example: ig-cli shell`,
	Run: func(cmd *cobra.Command, args []string) {
		startInteractiveMode()
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands (login/logout)",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Instagram",
	Run: func(cmd *cobra.Command, args []string) {
		auth := auth.NewInstagramAuth()
		client, err := auth.Login()
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully logged in as @%s\n", client.GetUsername())
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Instagram",
	Run: func(cmd *cobra.Command, args []string) {
		auth := auth.NewInstagramAuth()
		if err := auth.Logout(""); err != nil {
			fmt.Printf("Logout failed: %v\n", err)
			os.Exit(1)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Instagram CLI configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		cfg := config.GetInstance()
		value := cfg.Get(key, nil)
		if value != nil {
			fmt.Println(value)
		} else {
			fmt.Printf("Configuration key '%s' not found\n", key)
			os.Exit(1)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		cfg := config.GetInstance()
		if err := cfg.Set(key, value); err != nil {
			fmt.Printf("Failed to set config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s = %s\n", key, value)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetInstance()
		values := cfg.List()
		for _, kv := range values {
			fmt.Printf("%s = %v\n", kv.Key, kv.Value)
		}
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to default",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement config reset functionality
		fmt.Println("Configuration reset to default")
	},
}

func displayTitle() {
	fmt.Print(
		`   ██████╗  ██████╗   ██████╗ ██████╗  █████╗ ███╗   ███╗
  ██╔════╝ ██╔═══██╗ ██╔════╝ ██╔══██╗██╔══██╗████╗ ████║
  ██║  ███╗██║   ██║ ██║  ███╗██████╔╝███████║██╔████╔██║
  ██║   ██║██║   ██║ ██║   ██║██╔══██╗██╔══██║██║╚██╔╝██║
  ╚██████╔╝╚██████╔╝ ╚██████╔╝██║  ██║██║  ██║██║ ╚═╝ ██║
   ╚═════╝  ╚═════╝   ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝
`)
	fmt.Println("Dedicated to my CSC263 group.")
	fmt.Println()
	fmt.Println("Type 'help' to see available commands.")
	fmt.Println("Pro Tip: Use vim-motion ('k', 'j') to navigate chats and messages.")
	fmt.Printf("Version: %s\n", version)
}

func startInteractiveMode() {
	displayTitle()
	fmt.Println("\nInteractive mode activated")
	fmt.Println("Type commands directly (e.g., 'help', 'auth login', 'auth logout')")
	fmt.Println("Type 'exit' or 'quit' to close the CLI")
	fmt.Println("Type 'clear' to clear the screen")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("ig-cli> ")

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			fmt.Print("ig-cli> ")
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			os.Exit(0)
		}

		if input == "clear" {
			// Clear screen (Windows)
			fmt.Print("\033[H\033[2J")
			displayTitle()
			fmt.Print("ig-cli> ")
			continue
		}

		executeInteractiveCommand(input)
		fmt.Print("ig-cli> ")
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}
}

func executeInteractiveCommand(input string) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	command := args[0]
	args = args[1:]

	switch command {
	case "help":
		displayInteractiveHelp()
	case "version":
		fmt.Printf("InstagramCLI v%s\n", version)
	case "auth":
		handleAuthCommand(args)
	case "config":
		handleConfigCommand(args)
	case "clear":
		// Already handled above
		return
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type 'help' to see available commands")
	}
}

func displayInteractiveHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  help                    - Show this help message")
	fmt.Println("  version                 - Show version information")
	fmt.Println("  auth login              - Login to Instagram")
	fmt.Println("  auth logout             - Logout from Instagram")
	fmt.Println("  config get [key]        - Get configuration value")
	fmt.Println("  config set [key] [value] - Set configuration value")
	fmt.Println("  config list             - List all configuration values")
	fmt.Println("  config reset            - Reset configuration to default")
	fmt.Println("  clear                   - Clear the screen")
	fmt.Println("  exit/quit               - Exit the CLI")
	fmt.Println()
}

func handleAuthCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Auth subcommands: login, logout")
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "login":
		auth := auth.NewInstagramAuth()
		client, err := auth.Login()
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}
		fmt.Printf("Successfully logged in as @%s\n", client.GetUsername())
	case "logout":
		auth := auth.NewInstagramAuth()
		if err := auth.Logout(""); err != nil {
			fmt.Printf("Logout failed: %v\n", err)
			return
		}
		fmt.Println("Successfully logged out")
	default:
		fmt.Printf("Unknown auth subcommand: %s\n", subcommand)
		fmt.Println("Available auth subcommands: login, logout")
	}
}

func handleConfigCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Config subcommands: get, set, list, reset")
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: config get [key]")
			return
		}
		key := args[1]
		cfg := config.GetInstance()
		value := cfg.Get(key, nil)
		if value != nil {
			fmt.Println(value)
		} else {
			fmt.Printf("Configuration key '%s' not found\n", key)
		}
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: config set [key] [value]")
			return
		}
		key := args[1]
		value := args[2]
		cfg := config.GetInstance()
		if err := cfg.Set(key, value); err != nil {
			fmt.Printf("Failed to set config: %v\n", err)
			return
		}
		fmt.Printf("Set %s = %s\n", key, value)
	case "list":
		cfg := config.GetInstance()
		values := cfg.List()
		for _, kv := range values {
			fmt.Printf("%s = %v\n", kv.Key, kv.Value)
		}
	case "reset":
		// TODO: Implement config reset functionality
		fmt.Println("Configuration reset to default")
	default:
		fmt.Printf("Unknown config subcommand: %s\n", subcommand)
		fmt.Println("Available config subcommands: get, set, list, reset")
	}
}

func main() {
	// Add auth subcommands
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)

	// Add config subcommands
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configResetCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
