package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
}
