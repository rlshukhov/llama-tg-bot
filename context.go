package main

import (
	"sync"
)

type ChatMessage struct {
	Role string
	Text string
}

type ChatContext struct {
	messages []ChatMessage
	mu       sync.Mutex
}

var chatContexts = make(map[int64]*ChatContext)
var contextMu sync.Mutex

func addMessageToContext(chatID int64, text string, role string) {
	return

	//contextMu.Lock()
	//defer contextMu.Unlock()
	//
	//context, exists := chatContexts[chatID]
	//if !exists {
	//	context = &ChatContext{}
	//	chatContexts[chatID] = context
	//}
	//
	//context.mu.Lock()
	//defer context.mu.Unlock()
	//
	//context.messages = append(context.messages, ChatMessage{Role: role, Text: text})
	//
	//maxContextMessages, err := strconv.Atoi(os.Getenv("MAX_CONTEXT_MESSAGES"))
	//if err != nil || maxContextMessages <= 0 {
	//	maxContextMessages = 10
	//}
	//
	//if len(context.messages) > maxContextMessages {
	//	context.messages = context.messages[len(context.messages)-maxContextMessages:]
	//}
}

func getContextForChat(chatID int64) []ChatMessage {
	return []ChatMessage{}

	//contextMu.Lock()
	//defer contextMu.Unlock()
	//
	//context, exists := chatContexts[chatID]
	//if !exists {
	//	return nil
	//}
	//
	//context.mu.Lock()
	//defer context.mu.Unlock()
	//
	//return append([]ChatMessage{}, context.messages...)
}
