package common

const ChatServiceName = "ChatService"

type ChatMessage struct {
	From string
	Text string
}

type RegisterArgs struct {
	Name string
}

type RegisterReply struct {
	ClientID string
	History  []ChatMessage
	Name     string
}

type MessageArgs struct {
	ClientID string
	Message  string
}

type MessagePushArgs struct {
	ClientID string
}
