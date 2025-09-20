package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/abhi-praj/GoGram/proto/generated"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewInstagramServiceClient(conn)

	// Example usage
	fmt.Println("=== Instagram gRPC Client Example ===")
	fmt.Println()

	// Check auth status
	fmt.Println("1. Checking authentication status...")
	authStatus, err := client.GetAuthStatus(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("Error checking auth status: %v", err)
		fmt.Println("   Make sure the gRPC server is running with: ./ig-cli --grpc")
		return
	} else {
		fmt.Printf("   Logged in: %v\n", authStatus.IsLoggedIn)
		if authStatus.IsLoggedIn {
			fmt.Printf("   Username: %s\n", authStatus.Username)
			fmt.Printf("   Unread count: %d\n", authStatus.UnreadCount)
			fmt.Printf("   Notifications running: %v\n", authStatus.NotificationsRunning)
		}
	}
	fmt.Println()

	// If not logged in, attempt login (this would need real credentials)
	if !authStatus.IsLoggedIn {
		fmt.Println("2. Attempting login (demo - will fail without real credentials)...")
		loginResp, err := client.Login(context.Background(), &pb.LoginRequest{
			Username: "demo_username",
			Password: "demo_password",
		})
		if err != nil {
			log.Printf("   Login error: %v", err)
		} else {
			fmt.Printf("   Login success: %v\n", loginResp.Success)
			fmt.Printf("   Message: %s\n", loginResp.Message)
		}
		fmt.Println()
	}

	// Get chats
	fmt.Println("3. Getting chats...")
	chatsResp, err := client.GetChats(context.Background(), &pb.GetChatsRequest{
		Limit: 5,
	})
	if err != nil {
		log.Printf("   Error getting chats: %v", err)
	} else {
		fmt.Printf("   Found %d chats:\n", len(chatsResp.Chats))
		for i, chat := range chatsResp.Chats {
			fmt.Printf("   %d. %s (ID: %s, Unread: %d)\n", i+1, chat.Title, chat.InternalId, chat.UnreadCount)
			if chat.LastMessage != "" {
				fmt.Printf("      Last: %s\n", chat.LastMessage)
			}
		}
	}
	fmt.Println()

	// Configuration examples
	fmt.Println("4. Configuration operations...")

	// List all config
	configResp, err := client.ListConfig(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("   Error listing config: %v", err)
	} else {
		fmt.Printf("   Found %d configuration entries:\n", len(configResp.Configs))
		for _, cfg := range configResp.Configs {
			fmt.Printf("   %s = %s\n", cfg.Key, cfg.Value)
		}
	}
	fmt.Println()

	// Get specific config
	getConfigResp, err := client.GetConfig(context.Background(), &pb.GetConfigRequest{
		Key: "language",
	})
	if err != nil {
		log.Printf("   Error getting config: %v", err)
	} else {
		if getConfigResp.Found {
			fmt.Printf("   language = %s\n", getConfigResp.Value)
		} else {
			fmt.Println("   language config not found")
		}
	}
	fmt.Println()

	// Set config
	setConfigResp, err := client.SetConfig(context.Background(), &pb.SetConfigRequest{
		Key:   "test_key",
		Value: "test_value",
	})
	if err != nil {
		log.Printf("   Error setting config: %v", err)
	} else {
		fmt.Printf("   Set config success: %v, Message: %s\n", setConfigResp.Success, setConfigResp.Message)
	}
	fmt.Println()

	// Example of streaming (notification stream)
	fmt.Println("5. Starting notification stream (will run for 10 seconds)...")
	stream, err := client.StreamNotifications(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("   Error starting notification stream: %v", err)
	} else {
		go func() {
			for {
				notification, err := stream.Recv()
				if err != nil {
					log.Printf("   Stream error: %v", err)
					return
				}
				fmt.Printf("   📧 Notification: %s from %s in %s\n",
					notification.MessagePreview, notification.Sender, notification.ChatTitle)
			}
		}()

		// Let it run for 10 seconds
		time.Sleep(10 * time.Second)
		stream.CloseSend()
	}

	fmt.Println("=== Client demo completed ===")
}
