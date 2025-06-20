package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LiteLLMResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

// User session to store conversation history and selected model
type UserSession struct {
	Model    string
	Messages []Message
}

var (
	bot          *tgbotapi.BotAPI
	userSessions = make(map[int64]*UserSession)
	litellmURL   string
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Get configuration from environment variables
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	litellmURL = os.Getenv("LITELLM_URL")

	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	if litellmURL == "" {
		litellmURL = "http://localhost:4000" // Default LiteLLM URL
	}

	// Initialize bot
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Handle updates
	for update := range updates {
		if update.Message != nil {
			handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	userID := message.Chat.ID

	// Initialize user session if it doesn't exist
	if userSessions[userID] == nil {
		userSessions[userID] = &UserSession{
			Model:    "perplexity/sonar", // Default model
			Messages: []Message{},
		}
	}

	session := userSessions[userID]

	switch {
	case strings.HasPrefix(message.Text, "/start"):
		handleStart(userID)
	case strings.HasPrefix(message.Text, "/model"):
		handleModelSelection(userID)
	case strings.HasPrefix(message.Text, "/clear"):
		handleClear(userID)
	case strings.HasPrefix(message.Text, "/help"):
		handleHelp(userID)
	case strings.HasPrefix(message.Text, "/status"):
		handleStatus(userID)
	case strings.HasPrefix(message.Text, "/system_prompt"):
		handleSetSystemPrompt(userID, message.Text)
	default:
		// Regular chat message
		handleChat(userID, message.Text, session)
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	userID := callback.Message.Chat.ID

	model := callback.Data

	// Initialize session if it doesn't exist
	if userSessions[userID] == nil {
		userSessions[userID] = &UserSession{
			Messages: []Message{},
		}
	}

	userSessions[userID].Model = model

	// Answer callback query
	bot.Request(tgbotapi.NewCallback(callback.ID, "Model selected!"))

	// Send confirmation message
	msg := tgbotapi.NewMessage(userID, fmt.Sprintf("✅ Model changed to: %s", model))
	bot.Send(msg)

	// Delete the keyboard message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	bot.Request(deleteMsg)
}
