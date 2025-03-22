package handlers

import (
	"strings"
	"telegram-bot/db"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			switch strings.ToLower(update.Message.Command()) {
			case "start":
				HandleStart(bot, update.Message)
			case "createprofile":
				HandleStart(bot, update.Message)
			case "profile":
				profile, err := db.GetProfile(update.Message.From.ID)
				if err != nil || profile == nil || profile.Name == "" {
					SendMessage(bot, update.Message.Chat.ID, "Профиль не найден. Используйте /createprofile для создания анкеты.")
				} else {
					SendMessage(bot, update.Message.Chat.ID, utils.FormatProfile(profile))
				}
			case "deleteprofile":
				HandleDeleteProfile(bot, update.Message)
			case "help":
				HandleUserHelp(bot, update.Message)
			// другие команды: /setname, /setage,...
			default:
				SendMessage(bot, update.Message.Chat.ID, "Неизвестная команда. Попробуйте /help для списка доступных команд.")
			}
		} else {
			// Если сообщение не команда и пользователь находится в процессе регистрации:
			if state, exists := registrationStates[update.Message.From.ID]; exists {
				HandleRegistrationConversation(bot, update.Message, state)
			}
		}
	}
}
