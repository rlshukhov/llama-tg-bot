package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// Определяем новую структуру QueueItem
type QueueItem struct {
	Message   tgbotapi.Message
	SentMsgID int
}

var allowedChatIDs map[int64]bool

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	messageQueue := make(chan QueueItem, 100)
	var wg sync.WaitGroup

	workerCount, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil || workerCount <= 0 {
		workerCount = 1
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(bot, messageQueue, &wg)
	}

	// Инициализация allowedChatIDs
	allowedChatIDs = make(map[int64]bool)
	allowedChatIDsStr := os.Getenv("ALLOWED_CHAT_IDS")
	if allowedChatIDsStr != "" {
		for _, idStr := range strings.Split(allowedChatIDsStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				log.Printf("Ошибка при парсинге ID чата: %v", err)
				continue
			}
			allowedChatIDs[id] = true
		}
	}

	for update := range updates {
		if update.Message != nil {
			// Проверяем, разрешен ли этот чат
			if !allowedChatIDs[update.Message.Chat.ID] {
				log.Printf("Игнорирование сообщения из неразрешенного чата: %d", update.Message.Chat.ID)
				continue
			}

			// Отправляем "queued" сразу после получения сообщения
			reply := tgbotapi.NewMessage(update.Message.Chat.ID, "queued")
			reply.ReplyToMessageID = update.Message.MessageID // Добавляем эту строку
			sentMsg, err := bot.Send(reply)
			if err != nil {
				log.Println("Ошибка отправки сообщения:", err)
				continue
			}

			// Создаем новую структуру, содержащую оригинальное сообщение и ID отправленного сообщения
			queueItem := QueueItem{
				Message:   *update.Message,
				SentMsgID: sentMsg.MessageID,
			}

			// Отправляем структуру в очередь
			messageQueue <- queueItem
		}
	}

	close(messageQueue)
	wg.Wait()
}
