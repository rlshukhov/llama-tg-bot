package main

import (
	"log"
	"strings"
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

	preparedMessage := strings.TrimSpace(strings.ReplaceAll(msg.Text, "/start@rlsaibot", ""))
	responseChan := getLLAMAResponse(preparedMessage, context)

	var fullResponse string
	lastUpdate := time.Now()

	formatR1 := func(s string) string {
		if strings.Contains(s, "<think>") && !strings.Contains(s, "</think>") {
			s = s + "\n</think>"
		}

		return strings.ReplaceAll(strings.ReplaceAll(s, "<think>", "<blockquote expandable>"), "</think>", "</blockquote>")
	}

	for chunk := range responseChan {
		fullResponse += chunk
		if time.Since(lastUpdate) >= time.Second*5 {
			edit := tgbotapi.NewEditMessageText(msg.Chat.ID, sentMsgID, formatR1(fullResponse)+"\n\ngenerating...")
			edit.ParseMode = tgbotapi.ModeHTML
			_, err := bot.Send(edit)
			if err != nil {
				log.Println("Ошибка обновления сообщения", err)
			}
			lastUpdate = time.Now()
		}
	}

	edit := tgbotapi.NewEditMessageText(msg.Chat.ID, sentMsgID, formatR1(fullResponse))
	edit.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(edit)
	if err != nil {
		log.Println("Ошибка обновления сообщения", err)
	}

	addMessageToContext(msg.Chat.ID, preparedMessage, "user")
	addMessageToContext(msg.Chat.ID, formatR1(fullResponse), "assistant")
}
