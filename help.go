package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleStart(userID int64) {
	msg := tgbotapi.NewMessage(userID,
		"🤖 Welcome to LiteLLM Bot!\n\n"+
			"I can connect you to various AI models through your LiteLLM server.\n\n"+
			"Commands:\n"+
			"/model - Select AI model\n"+
			"/clear - Clear conversation history\n"+
			"/status - Show current model and message count\n"+
			"/help - Show this help message\n\n"+
			"Just send me a message to start chatting!")

	bot.Send(msg)
}

func handleHelp(userID int64) {
	helpText := "🤖 LiteLLM Bot Help\n\n" +
		"Commands:\n" +
		"/start - Start the bot\n" +
		"/model - Change AI model\n" +
		"/clear - Clear conversation history\n" +
		"/status - Show current settings\n" +
		"/help - Show this help\n\n" +
		"Just send any message to chat with the AI!"

	msg := tgbotapi.NewMessage(userID, helpText)
	bot.Send(msg)
}

func handleStatus(userID int64) {
	session := userSessions[userID]
	if session == nil {
		session = &UserSession{
			Model:    "gpt-3.5-turbo",
			Messages: []Message{},
		}
		userSessions[userID] = session
	}

	statusText := fmt.Sprintf(
		"📊 Current Status:\n\n"+
			"🧠 Model: %s\n"+
			"💬 Messages in history: %d\n"+
			"🔗 LiteLLM Server: %s",
		session.Model,
		len(session.Messages),
		litellmURL,
	)

	msg := tgbotapi.NewMessage(userID, statusText)
	bot.Send(msg)
}
