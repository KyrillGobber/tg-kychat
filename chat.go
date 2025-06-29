package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LiteLLMResponseFull struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Citations []string `json:"citations,omitempty"` // Optional field for citations
}

func handleClear(userID int64) {
	if userSessions[userID] != nil {
		userSessions[userID].Messages = []Message{}
	}

	setSystemPrompt(userID, userSessions[userID].SystemPrompt)

	msg := tgbotapi.NewMessage(userID, "🗑️ Conversation history cleared!")
	bot.Send(msg)
}

func handleChat(userID int64, text string, session *UserSession) {
	// Add user message to history
	session.Messages = append(session.Messages, Message{
		Role:    "user",
		Content: text,
	})

	// Send typing indicator
	typing := tgbotapi.NewChatAction(userID, tgbotapi.ChatTyping)
	bot.Send(typing)
	// ... message
	generatingMessage := tgbotapi.NewMessage(userID, "...")
	sentMsg, _ := bot.Send(generatingMessage)

	// Make request to LiteLLM
	response, err := callLiteLLM(session.Model, session.Messages)
	if err != nil {
		log.Printf("Error calling LiteLLM: %v", err)
		msg := tgbotapi.NewMessage(userID,
			"❌ Sorry, I encountered an error while processing your request. "+
				"Please check if the LiteLLM server is running and accessible.")
		bot.Send(msg)
		return
	}

	// Add assistant response to history
	assistantMessage := response.Choices[0].Message
	tokenUsage := fmt.Sprintf("Tokens: %d (prompt: %d, completion: %d)",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens)
	session.Messages = append(session.Messages, assistantMessage)

	msgText := fmt.Sprintf("%s\n---\n%s", assistantMessage.Content, tokenUsage)

	// Edit the generating message with the actual response
	finalMsg := tgbotapi.NewEditMessageText(userID, sentMsg.MessageID, msgText)
	finalMsg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(finalMsg)
	if len(response.Citations) > 0 {
		citationMsg := tgbotapi.NewMessage(userID, "Citations:\n"+formatCitations(response.Citations))
		citationMsg.ParseMode = tgbotapi.ModeMarkdown
		bot.Send(citationMsg)
	}
}

func formatCitations(citations []string) string {
	if len(citations) == 0 {
		return "No citations available."
	}

	formatted := ""
	for i, citation := range citations {
		formatted += fmt.Sprintf("%d. %s\n", i+1, citation)
	}
	return formatted
}
