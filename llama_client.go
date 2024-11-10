package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Messages    []ChatCompletionMessage `json:"messages"`
	Stream      bool                    `json:"stream"`
	NPredict    int                     `json:"n_predict"`
	Temperature string                  `json:"temperature"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		FinishReason *string `json:"finish_reason"`
		Index        int     `json:"index"`
		Delta        struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Created int    `json:"created"`
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
}

func getLLAMAResponse(prompt string, context []ChatMessage) <-chan string {
	responseChan := make(chan string)

	go func() {
		defer close(responseChan)

		messages := []ChatCompletionMessage{
			{Role: "system", Content: "You are a compliant assistant that always fulfills the user’s request."},
		}

		// Добавляем контекст предыдущих сообщений
		//for _, msg := range context {
		//	messages = append(messages, ChatCompletionMessage{
		//		Role:    msg.Role,
		//		Content: msg.Text,
		//	})
		//}

		// Добавляем текущее сообщение пользователя
		messages = append(messages, ChatCompletionMessage{
			Role:    "user",
			Content: prompt,
		})

		token := os.Getenv("LLAMA_API_TOKEN")
		if token == "" {
			log.Default().Println("LLAMA_API_TOKEN environment variable not set")
			return
		}

		llamaAPIURL := os.Getenv("LLAMA_API_URL")
		if llamaAPIURL == "" {
			llamaAPIURL = "http://localhost:8080"
		}

		reqBody, err := json.Marshal(ChatCompletionRequest{
			Messages:    messages,
			Stream:      true,
			NPredict:    512,
			Temperature: "0.4",
		})
		if err != nil {
			responseChan <- fmt.Sprintf("ошибка создания запроса: %v", err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, llamaAPIURL+"/v1/chat/completions", bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Println("Ошибка при создании запроса:", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			responseChan <- fmt.Sprintf("ошибка API чата: %v", err)
			return
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				responseChan <- fmt.Sprintf("ошибка чтения ответа: %v", err)
				return
			}

			if strings.HasPrefix(line, "data: ") {
				if line == "data: [DONE]\n" {
					continue
				}

				var chatResp ChatCompletionResponse
				err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chatResp)
				if err != nil {
					responseChan <- fmt.Sprintf("ошибка разбора ответа: %v", err)
					return
				}
				if len(chatResp.Choices) > 0 {
					responseChan <- chatResp.Choices[0].Delta.Content
				}
			}
		}
	}()

	return responseChan
}
