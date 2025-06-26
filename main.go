package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
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
}

type Choice struct {
	Message Message `json:"message"`
}

type UserSession struct {
	Model        string
	Messages     []Message
	SystemPrompt string
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
	allowedUsers := strings.Split(os.Getenv("ALLOWED_USERS"), ",")

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
		// Only allow allowed users
		if update.Message != nil && slices.Contains(allowedUsers, strconv.FormatInt(update.Message.From.ID, 10)) {
			handleMessage(update.Message)
		} else if update.CallbackQuery != nil && slices.Contains(allowedUsers, strconv.FormatInt(update.CallbackQuery.From.ID, 10)) {
			handleCallbackQuery(update.CallbackQuery)
		} else {
			userID := update.Message.From.ID
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf("Sorry %s, you can't use this bot", update.Message.From.UserName))
			bot.Send(msg)
			log.Printf("User %s tried to access the bot %d", update.Message.From.UserName, userID)
			continue
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	userID := message.Chat.ID

	// Initialize user session if it doesn't exist
	if userSessions[userID] == nil {
		userSessions[userID] = &UserSession{
			Model:        "mistral-31-24b", // Default model
			Messages:     []Message{},
			SystemPrompt: "You are my bruv, be witty and concise.",
		}
	}

	session := userSessions[userID]

	switch {
	case strings.HasPrefix(message.Text, "/model"):
		handleModelSelection(userID)
	case strings.HasPrefix(message.Text, "/system_prompt"):
		handleSetSystemPrompt(userID, message.Text)
	case strings.HasPrefix(message.Text, "/clear"):
		handleClear(userID)
	case strings.HasPrefix(message.Text, "/status"):
		handleStatus(userID)
	case strings.HasPrefix(message.Text, "/help"):
		handleHelp(userID)
	default:
		// Regular chat message
		handleChat(userID, message.Text, session)
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	userID := callback.Message.Chat.ID

	model := callback.Data

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

func handleSetSystemPrompt(userID int64, text string) {
	// Extract the system prompt from the command
	prompt := strings.TrimPrefix(text, "/system_prompt")

	if prompt == "" {
		msg := tgbotapi.NewMessage(userID, "Use like /system_prompt <your prompt> to set a system prompt.")
		bot.Send(msg)
		return
	}

	setSystemPrompt(userID, prompt)

	msg := tgbotapi.NewMessage(userID, "System prompt set successfully!")
	bot.Send(msg)
}

func setSystemPrompt(userID int64, prompt string) {
	// Set the system prompt in the user's session
	session := userSessions[userID]
	session.SystemPrompt = prompt
	session.Messages = append(session.Messages, Message{
		Role:    "system",
		Content: prompt,
	})
}
