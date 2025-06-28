package main

import (
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type LiteLLMResp struct {
	Data   []ModelInfo `json:"data"`
	Object string      `json:"object"`
}

// Create inline keyboard with model options
func handleModelSelection(userID int64) {
	var msg tgbotapi.MessageConfig
	models, err := getLiteLLMModels()
	if err != nil {
		msg = tgbotapi.NewMessage(userID, fmt.Sprintf("Error fetching models: %v", err))
	} else {
		sort.Slice(models, func(i, j int) bool {
			return models[i].ID < models[j].ID
		})
		keyboard := tgbotapi.NewInlineKeyboardMarkup()
		for _, model := range models {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
				tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(model.ID, model.ID)))
		}
		msg = tgbotapi.NewMessage(userID, "Choose your AI model:")
		msg.ReplyMarkup = keyboard
	}
	bot.Send(msg)
}
