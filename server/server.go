package main

import (
	"chat-by-rpc/common"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// clientChannel is a channel used to push messages to a single client.
type clientChannel chan common.ChatMessage

// ChatService holds the state of the chat server.
type ChatService struct {
	mu      sync.Mutex
	clients map[string]string    // clientID → name
	history []common.ChatMessage // full chat log

	// Channel for all new messages sent to the server.
	inboundMessages chan common.ChatMessage

	// Map to hold a channel for each client to push new messages.
	// clientID -> clientChannel
	clientChannels map[string]clientChannel
}

func NewChatService() *ChatService {
	cs := &ChatService{
		clients:         make(map[string]string),
		history:         []common.ChatMessage{},
		inboundMessages: make(chan common.ChatMessage, 10), // Buffered global message channel
		clientChannels:  make(map[string]clientChannel),
	}
	go cs.pusher() // Start the message distribution goroutine
	return cs
}

// pusher listens on the inboundMessages channel and sends the message to all clients.
func (s *ChatService) pusher() {
	for msg := range s.inboundMessages {
		s.mu.Lock()

		// 1. Add to history
		s.history = append(s.history, msg)

		// 2. Distribute to all client channels
		for _, ch := range s.clientChannels {
			// Non-blocking send to avoid locking up the pusher if a client's channel is full/dead
			select {
			case ch <- msg:
				// Message sent successfully
			default:
				// Channel full, likely a slow/dead client.
				log.Printf("Warning: Client channel full, dropping message: %v", msg)
			}
		}
		s.mu.Unlock()
		fmt.Printf("[Pusher] Sent to %d clients\n", len(s.clientChannels))
	}
}

func (s *ChatService) RegisterClient(args common.RegisterArgs, reply *common.RegisterReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("C%d", len(s.clients)+1)
	s.clients[id] = args.Name
	s.clientChannels[id] = make(clientChannel, 5) // Create a new channel for the client

	// Join message
	joinMsg := common.ChatMessage{
		From: "SYSTEM",
		Text: fmt.Sprintf("%s joined the chat", args.Name),
	}
	s.inboundMessages <- joinMsg // Send join message via the channel

	reply.ClientID = id
	reply.History = append([]common.ChatMessage{}, s.history...) // Send a copy of the current history
	reply.Name = args.Name

	fmt.Println("User joined:", args.Name, "with ID:", id)
	return nil
}

// SendMessage sends the message to the global inboundMessages channel for distribution.
func (s *ChatService) SendMessage(args common.MessageArgs, _ *struct{}) error {
	s.mu.Lock()
	sender := s.clients[args.ClientID]
	s.mu.Unlock()

	if sender == "" {
		return fmt.Errorf("unknown client ID: %s", args.ClientID)
	}

	msg := common.ChatMessage{
		From: sender,
		Text: args.Message,
	}

	// Send to the inbound channel to be processed by the pusher goroutine
	s.inboundMessages <- msg

	fmt.Printf("[RPC] Sent message from [%s] to inbound channel\n", sender)
	return nil
}

// WaitForMessage is a blocking RPC call that waits on the client's dedicated channel.
func (s *ChatService) WaitForMessage(args common.MessagePushArgs, reply *common.ChatMessage) error {
	s.mu.Lock()
	ch, ok := s.clientChannels[args.ClientID]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("client ID %s not registered", args.ClientID)
	}

	// Wait for a new message on the client's channel
	// use a timeout to prevent an indefinite hang if the connection breaks or server shuts down.
	select {
	case msg := <-ch:
		*reply = msg
		return nil
	case <-time.After(5 * time.Second): // Timeout after 5 seconds
		// Return a harmless error to let the client retry the call
		return fmt.Errorf("timeout waiting for new message")
	}
}

func main() {
	chat := NewChatService()
	rpc.RegisterName(common.ChatServiceName, chat)

	ln, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server running on 127.0.0.1:1234")
	for {
		conn, err := ln.Accept()
		if err == nil {
			go rpc.ServeConn(conn)
		} else {
			log.Println("Accept error:", err)
		}
	}
}
