package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/abhi-praj/GoGram/internal/auth"
	"github.com/abhi-praj/GoGram/internal/chat"
	"github.com/abhi-praj/GoGram/internal/client"
	"github.com/abhi-praj/GoGram/internal/config"
	pb "github.com/abhi-praj/GoGram/proto/generated"
)

// Server implements the InstagramService gRPC server
type Server struct {
	pb.UnimplementedInstagramServiceServer
	authInstance   *auth.InstagramAuth
	clientInstance *client.ClientWrapper
	dmInstance     *chat.DirectMessages
	config         *config.Config

	// Streaming connections
	messageStreams map[string][]pb.InstagramService_StreamMessagesServer
	notifStreams   []pb.InstagramService_StreamNotificationsServer
	streamMutex    sync.RWMutex

	// Server control
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer creates a new gRPC server instance
func NewServer() *Server {
	return &Server{
		authInstance:   auth.NewInstagramAuth(),
		config:         config.GetInstance(),
		messageStreams: make(map[string][]pb.InstagramService_StreamMessagesServer),
		notifStreams:   make([]pb.InstagramService_StreamNotificationsServer, 0),
	}
}

// Start starts the gRPC server on the specified address
func (s *Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.listener = lis
	s.grpcServer = grpc.NewServer()
	pb.RegisterInstagramServiceServer(s.grpcServer, s)

	log.Printf("gRPC server starting on %s", address)
	return s.grpcServer.Serve(lis)
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	if s.grpcServer != nil {
		log.Println("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
	}
}

// Authentication methods

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return &pb.LoginResponse{
			Success: false,
			Message: "Username and password are required",
		}, nil
	}

	// Create a new client wrapper for this login attempt
	clientWrapper := client.NewClientWrapper(req.Username)

	// Attempt login
	if err := clientWrapper.Login(req.Username, req.Password, req.VerificationCode); err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: fmt.Sprintf("Login failed: %v", err),
		}, nil
	}

	// Store successful login
	s.clientInstance = clientWrapper
	s.dmInstance = chat.NewDirectMessages(clientWrapper)

	// Start notifications
	if err := s.dmInstance.StartNotifications(); err != nil {
		log.Printf("Warning: Could not start notifications: %v", err)
	}

	return &pb.LoginResponse{
		Success:  true,
		Message:  "Login successful",
		Username: req.Username,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if s.clientInstance == nil {
		return &pb.LogoutResponse{
			Success: false,
			Message: "Not logged in",
		}, nil
	}

	// Stop notifications
	if s.dmInstance != nil {
		s.dmInstance.StopNotifications()
	}

	// Logout
	if err := s.authInstance.Logout(req.Username); err != nil {
		return &pb.LogoutResponse{
			Success: false,
			Message: fmt.Sprintf("Logout failed: %v", err),
		}, nil
	}

	s.clientInstance = nil
	s.dmInstance = nil

	return &pb.LogoutResponse{
		Success: true,
		Message: "Logout successful",
	}, nil
}

func (s *Server) GetAuthStatus(ctx context.Context, req *emptypb.Empty) (*pb.AuthStatusResponse, error) {
	if s.clientInstance == nil {
		return &pb.AuthStatusResponse{
			IsLoggedIn: false,
		}, nil
	}

	response := &pb.AuthStatusResponse{
		IsLoggedIn: true,
		Username:   s.clientInstance.GetUsername(),
	}

	if s.dmInstance != nil {
		if count, err := s.dmInstance.GetUnreadCount(); err == nil {
			response.UnreadCount = int32(count)
		}
		response.NotificationsRunning = s.dmInstance.IsNotificationRunning()
	}

	return response, nil
}

// Chat methods

func (s *Server) GetChats(ctx context.Context, req *pb.GetChatsRequest) (*pb.GetChatsResponse, error) {
	if s.dmInstance == nil {
		return nil, status.Error(codes.Unauthenticated, "Not logged in")
	}

	var chats []*chat.Chat
	var err error

	if req.Limit > 0 {
		chats, err = s.dmInstance.GetChatsWithLimit(int(req.Limit))
	} else {
		chats, err = s.dmInstance.GetChats()
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get chats: %v", err)
	}

	// Convert to protobuf format
	pbChats := make([]*pb.Chat, len(chats))
	for i, chat := range chats {
		pbChats[i] = s.convertChatToPB(chat)
	}

	return &pb.GetChatsResponse{
		Chats:      pbChats,
		TotalCount: int32(len(chats)),
	}, nil
}

func (s *Server) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	if s.dmInstance == nil {
		return nil, status.Error(codes.Unauthenticated, "Not logged in")
	}

	messages, err := s.dmInstance.GetChatHistory(req.ChatId, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get messages: %v", err)
	}

	// Convert to protobuf format
	pbMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		pbMessages[i] = s.convertMessageToPB(msg)
	}

	return &pb.GetMessagesResponse{
		Messages: pbMessages,
		HasMore:  len(messages) == int(req.Limit), // Simple heuristic
	}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	if s.dmInstance == nil {
		return nil, status.Error(codes.Unauthenticated, "Not logged in")
	}

	err := s.dmInstance.SendMessageByInternalID(req.ChatId, req.Message)
	if err != nil {
		return &pb.SendMessageResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.SendMessageResponse{
		Success:   true,
		MessageId: fmt.Sprintf("msg_%d", time.Now().Unix()), // Simple ID generation
	}, nil
}

func (s *Server) StartInteractiveChat(ctx context.Context, req *pb.StartInteractiveChatRequest) (*pb.StartInteractiveChatResponse, error) {
	if s.dmInstance == nil {
		return nil, status.Error(codes.Unauthenticated, "Not logged in")
	}

	err := s.dmInstance.StartInteractiveChat(req.ChatId)
	if err != nil {
		return &pb.StartInteractiveChatResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start interactive chat: %v", err),
		}, nil
	}

	return &pb.StartInteractiveChatResponse{
		Success: true,
		Message: "Interactive chat started",
	}, nil
}

// Streaming methods

func (s *Server) StreamMessages(req *pb.StreamMessagesRequest, stream pb.InstagramService_StreamMessagesServer) error {
	if s.dmInstance == nil {
		return status.Error(codes.Unauthenticated, "Not logged in")
	}

	// Add this stream to the list for the chat
	s.streamMutex.Lock()
	if s.messageStreams[req.ChatId] == nil {
		s.messageStreams[req.ChatId] = make([]pb.InstagramService_StreamMessagesServer, 0)
	}
	s.messageStreams[req.ChatId] = append(s.messageStreams[req.ChatId], stream)
	s.streamMutex.Unlock()

	// Keep the stream alive
	<-stream.Context().Done()

	// Remove stream when done
	s.streamMutex.Lock()
	streams := s.messageStreams[req.ChatId]
	for i, streamItem := range streams {
		if streamItem == stream {
			s.messageStreams[req.ChatId] = append(streams[:i], streams[i+1:]...)
			break
		}
	}
	s.streamMutex.Unlock()

	return nil
}

func (s *Server) StreamNotifications(req *emptypb.Empty, stream pb.InstagramService_StreamNotificationsServer) error {
	if s.dmInstance == nil {
		return status.Error(codes.Unauthenticated, "Not logged in")
	}

	// Add this stream to the notification streams
	s.streamMutex.Lock()
	s.notifStreams = append(s.notifStreams, stream)
	s.streamMutex.Unlock()

	// Keep the stream alive
	<-stream.Context().Done()

	// Remove stream when done
	s.streamMutex.Lock()
	for i, streamItem := range s.notifStreams {
		if streamItem == stream {
			s.notifStreams = append(s.notifStreams[:i], s.notifStreams[i+1:]...)
			break
		}
	}
	s.streamMutex.Unlock()

	return nil
}

// Configuration methods

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	value := s.config.Get(req.Key, nil)

	response := &pb.GetConfigResponse{
		Key:   req.Key,
		Found: value != nil,
	}

	if value != nil {
		response.Value = fmt.Sprintf("%v", value)
	}

	return response, nil
}

func (s *Server) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
	err := s.config.Set(req.Key, req.Value)
	if err != nil {
		return &pb.SetConfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to set config: %v", err),
		}, nil
	}

	return &pb.SetConfigResponse{
		Success: true,
		Message: "Configuration updated",
	}, nil
}

func (s *Server) ListConfig(ctx context.Context, req *emptypb.Empty) (*pb.ListConfigResponse, error) {
	values := s.config.List()

	configs := make([]*pb.ConfigKeyValue, len(values))
	for i, kv := range values {
		configs[i] = &pb.ConfigKeyValue{
			Key:   kv.Key,
			Value: fmt.Sprintf("%v", kv.Value),
		}
	}

	return &pb.ListConfigResponse{
		Configs: configs,
	}, nil
}

// Helper methods to convert between internal types and protobuf types

func (s *Server) convertChatToPB(chat *chat.Chat) *pb.Chat {
	pbChat := &pb.Chat{
		Id:          chat.ID,
		InternalId:  chat.InternalID,
		Title:       chat.Title,
		LastMessage: chat.LastMessage,
		UnreadCount: int32(chat.UnreadCount),
		IsGroup:     chat.IsGroup,
	}

	if !chat.LastActivity.IsZero() {
		pbChat.LastActivity = timestamppb.New(chat.LastActivity)
	}

	// Convert users
	pbChat.Users = make([]*pb.User, len(chat.Users))
	for i, user := range chat.Users {
		pbChat.Users[i] = &pb.User{
			Id:            fmt.Sprintf("%d", user.ID),
			Username:      user.Username,
			FullName:      user.FullName,
			ProfilePicUrl: user.ProfilePicURL,
			IsVerified:    user.IsVerified,
		}
	}

	return pbChat
}

func (s *Server) convertMessageToPB(msg *chat.Message) *pb.Message {
	pbMsg := &pb.Message{
		Id:     msg.ID,
		Text:   msg.Text,
		Sender: msg.Sender,
		Type:   pb.MessageType_TEXT, // Default to text
	}

	if !msg.Timestamp.IsZero() {
		pbMsg.Timestamp = timestamppb.New(msg.Timestamp)
	}

	// Map message types
	switch msg.Type {
	case "text":
		pbMsg.Type = pb.MessageType_TEXT
	case "media":
		pbMsg.Type = pb.MessageType_MEDIA
	case "system":
		pbMsg.Type = pb.MessageType_SYSTEM
	}

	return pbMsg
}

// BroadcastMessageUpdate sends a message update to all connected streams for a chat
func (s *Server) BroadcastMessageUpdate(chatID string, message *chat.Message, updateType pb.MessageUpdateType) {
	s.streamMutex.RLock()
	streams := s.messageStreams[chatID]
	s.streamMutex.RUnlock()

	if len(streams) == 0 {
		return
	}

	update := &pb.MessageUpdate{
		ChatId:  chatID,
		Message: s.convertMessageToPB(message),
		Type:    updateType,
	}

	// Send to all streams for this chat
	for _, stream := range streams {
		if err := stream.Send(update); err != nil {
			log.Printf("Error sending message update: %v", err)
		}
	}
}

// BroadcastNotification sends a notification to all connected notification streams
func (s *Server) BroadcastNotification(chatID, chatTitle, sender, messagePreview string, unreadCount int) {
	s.streamMutex.RLock()
	streams := s.notifStreams
	s.streamMutex.RUnlock()

	if len(streams) == 0 {
		return
	}

	notification := &pb.NotificationUpdate{
		ChatId:         chatID,
		ChatTitle:      chatTitle,
		Sender:         sender,
		MessagePreview: messagePreview,
		Timestamp:      timestamppb.New(time.Now()),
		UnreadCount:    int32(unreadCount),
	}

	// Send to all notification streams
	for _, stream := range streams {
		if err := stream.Send(notification); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}
}
