package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleHelp(userID int64) {
	helpText := "🤖 LiteLLM Bot Help\n\n" +
		"Commands:\n" +
		"/model - Change AI model\n" +
		"/system_prompt - Change system prompt\n" +
		"/clear - Clear conversation history\n" +
		"/status - Show current settings\n" +
		"/help - Show this help\n\n" +
		"Just send any message to chat with the default model."

	msg := tgbotapi.NewMessage(userID, helpText)
	bot.Send(msg)
}

func handleStatus(userID int64) {
	session := userSessions[userID]

	statusText := fmt.Sprintf(
		"📊 Current Status:\n\n"+
			"🧠 Model: %s\n"+
			"🤖 System Prompt: %s\n"+
			"💬 Messages in history: %d (includes system prompt)\n"+
			"🔗 LiteLLM Server: %s",
		session.Model,
		session.SystemPrompt,
		len(session.Messages),
		litellmURL,
	)

	msg := tgbotapi.NewMessage(userID, statusText)
	bot.Send(msg)
}
