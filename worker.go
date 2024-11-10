package main

import (
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func worker(bot *tgbotapi.BotAPI, messageQueue <-chan QueueItem, wg *sync.WaitGroup) {
	defer wg.Done()

	for item := range messageQueue {
		// Дополнительная проверка разрешенных чатов
		if !allowedChatIDs[item.Message.Chat.ID] {
			log.Printf("Игнорирование сообщения из неразрешенного чата: %d", item.Message.Chat.ID)
			continue
		}

		// Обновляем сообщение на "generating" перед началом обработки
		edit := tgbotapi.NewEditMessageText(item.Message.Chat.ID, item.SentMsgID, "generating")
		_, err := bot.Send(edit)
		if err != nil {
			log.Println("Ошибка обновления сообщения", err)
		}

		processMessage(bot, item.Message, item.SentMsgID)
	}
}

func processMessage(bot *tgbotapi.BotAPI, msg tgbotapi.Message, sentMsgID int) {
	context := getContextForChat(msg.Chat.ID)

	responseChan := getLLAMAResponse(msg.Text, context)

	var fullResponse string
	lastUpdate := time.Now()

	for chunk := range responseChan {
		fullResponse += chunk
		if time.Since(lastUpdate) >= time.Second*5 {
			edit := tgbotapi.NewEditMessageText(msg.Chat.ID, sentMsgID, fullResponse+"\n\ngenerating...")
			edit.ParseMode = tgbotapi.ModeMarkdown
			_, err := bot.Send(edit)
			if err != nil {
				log.Println("Ошибка обновления сообщения", err)
			}
			lastUpdate = time.Now()
		}
	}

	edit := tgbotapi.NewEditMessageText(msg.Chat.ID, sentMsgID, fullResponse)
	edit.ParseMode = tgbotapi.ModeMarkdown
	_, err := bot.Send(edit)
	if err != nil {
		log.Println("Ошибка обновления сообщения", err)
	}

	addMessageToContext(msg.Chat.ID, msg.Text, "user")
	addMessageToContext(msg.Chat.ID, fullResponse, "assistant")
}
