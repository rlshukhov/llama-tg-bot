package main

import (
	"github.com/microcosm-cc/bluemonday"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func mdToHTML(md string) string {
	extensions := parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	doc := p.Parse([]byte(md))
	h := string(markdown.Render(doc, renderer))

	policy := bluemonday.NewPolicy()
	policy.AllowAttrs("expandable").OnElements("blockquote")
	policy.AllowAttrs("href").OnElements("a")
	policy.AllowElements("b", "strong", "i", "em", "code", "s", "strike", "del", "u", "pre")

	return policy.Sanitize(h)
}

func worker(bot *tgbotapi.BotAPI, messageQueue <-chan QueueItem, wg *sync.WaitGroup) {
	defer wg.Done()

	for item := range messageQueue {
		if !allowedChatIDs[item.Message.Chat.ID] {
			log.Printf("Игнорирование сообщения из неразрешенного чата: %d", item.Message.Chat.ID)
			continue
		}

		edit := tgbotapi.NewEditMessageText(item.Message.Chat.ID, item.SentMsgID, "generating")
		_, err := bot.Send(edit)
		if err != nil {
			log.Println("Ошибка обновления сообщения", err)
		}

		processMessage(bot, item.Message, item.SentMsgID)
	}
}

func processMessage(bot *tgbotapi.BotAPI, msg tgbotapi.Message, sentMsgID int) {
	preparedMessage := strings.TrimSpace(strings.ReplaceAll(msg.Text, "/start@rlsaibot", ""))
	responseChan := getLLAMAResponse(preparedMessage)

	var fullResponse string
	lastUpdate := time.Now()

	formatR1 := func(s string) string {
		if strings.Contains(s, "<think>") && !strings.Contains(s, "</think>") {
			s = s + "\n</think>"
		}

		return mdToHTML(
			strings.ReplaceAll(
				strings.ReplaceAll(
					s,
					"<think>",
					"<blockquote expandable>",
				),
				"</think>",
				"</blockquote>",
			),
		)
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
}
