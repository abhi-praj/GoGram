package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abhi-praj/GoGram/internal/auth"
	"github.com/abhi-praj/GoGram/internal/chat"
	"github.com/abhi-praj/GoGram/internal/client"
	"github.com/rivo/tview"
)

var (
	version        = "0.1.0"
	authInstance   *auth.InstagramAuth
	clientInstance *client.ClientWrapper
	dmInstance     *chat.DirectMessages
)

func main() {
	// Initialize auth
	authInstance = auth.NewInstagramAuth()

	// Check if user is already logged in
	client, err := authInstance.Login()
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		fmt.Println("Please check your credentials and try again.")
		os.Exit(1)
	}

	clientInstance = client
	dmInstance = chat.NewDirectMessages(client)

	// Start the TUI
	if err := startTUI(); err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}

// startTUI initializes and runs the Terminal User Interface
func startTUI() error {
	// Create the tview application
	app := tview.NewApplication()

	// Create the chat interface
	chatInterface := chat.NewChatInterface(
		app,
		func(chatID, message string) error {
			// Handle message sending
			return dmInstance.SendMessageByInternalID(chatID, message)
		},
		func(chatID, message, replyToID string) error {
			// Handle reply sending - implement when reply functionality is available
			return fmt.Errorf("reply functionality not yet implemented")
		},
		func(messageID string) error {
			// Handle message unsending - implement when unsend functionality is available
			return fmt.Errorf("unsend functionality not yet implemented")
		},
		dmInstance,
	)

	// Load chats into the interface
	chats, err := dmInstance.GetChats()
	if err != nil {
		return fmt.Errorf("failed to load chats: %v", err)
	}

	// Set chats in the interface
	chatInterface.SetChats(chats)

	// Run the interface
	return chatInterface.Run()
}
