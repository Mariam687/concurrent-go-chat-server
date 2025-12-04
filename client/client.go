package main

import (
	"bufio"
	"chat-by-rpc/common"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
)

func main() {
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("Could not connect to server:", err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Register with server
	var reg common.RegisterReply
	err = client.Call(common.ChatServiceName+".RegisterClient",
		common.RegisterArgs{Name: name}, &reg)
	if err != nil {
		log.Fatal("Registration failed:", err)
	}

	fmt.Printf("Connected as **%s** (ID: %s)\n", reg.Name, reg.ClientID)

	// Display initial history
	fmt.Println("\n---------------")
	for _, msg := range reg.History {
		fmt.Printf("[%s] %s\n", msg.From, msg.Text)
	}
	fmt.Println("----------------")

	// Background message-receiving goroutine
	go func() {
		args := common.MessagePushArgs{ClientID: reg.ClientID}

		// The client will constantly call WaitForMessage
		// and the server will block until a message is ready.
		for {
			var msg common.ChatMessage
			err := client.Call(common.ChatServiceName+".WaitForMessage", args, &msg)

			if err != nil {
				continue // Retry the call
			}

			// Process received message
			// Ignore my own messages
			if msg.From == reg.Name {
				continue
			}

			// Print the new message
			fmt.Printf("\n📢 [%s] %s\n> ", msg.From, msg.Text)
		}
	}()

	// Main input loop
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "exit" {
			fmt.Println("Goodbye.")
			return
		}

		if text == "" {
			continue
		}

		// Send message RPC
		err = client.Call(common.ChatServiceName+".SendMessage",
			common.MessageArgs{
				ClientID: reg.ClientID,
				Message:  text,
			}, nil)

		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
