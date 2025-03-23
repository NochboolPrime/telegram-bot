package handlers

import (
	"strconv"
	"strings"
	"telegram-bot/db"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleUpdate обрабатывает входящие обновления для пользовательского бота.
func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.IsCommand() {
			switch strings.ToLower(update.Message.Command()) {
			case "start":
				HandleStart(bot, update.Message)
			case "createprofile":
				HandleStart(bot, update.Message) // Если профиль уже существует, HandleStart уведомит.
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
			case "attend":
				HandleAttendEvent(bot, update.Message)
			case "unattend":
				HandleUnattendEvent(bot, update.Message)
			// Можно добавить другие команды (например, /setname, /setage, и т.д.)
			default:
				SendMessage(bot, update.Message.Chat.ID, "Неизвестная команда. Попробуйте /help для списка доступных команд.")
			}
		} else {
			// Если сообщение не команда и пользователь находится в процессе регистрации,
			// обрабатываем последовательность регистрации.
			if state, exists := registrationStates[update.Message.From.ID]; exists {
				HandleRegistrationConversation(bot, update.Message, state)
			}
		}
	}
}

// HandleAttendEvent обрабатывает команду /attend <event_id>.
func HandleAttendEvent(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	arg := strings.TrimSpace(msg.CommandArguments())
	if arg == "" {
		SendMessage(bot, msg.Chat.ID, "Используйте: /attend <event_id>")
		return
	}
	eventID, err := strconv.Atoi(arg)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Event ID должно быть числом.")
		return
	}
	event, err := db.GetEventByID(eventID)
	if err != nil || event == nil {
		SendMessage(bot, msg.Chat.ID, "Событие не найдено.")
		return
	}
	if !event.Active {
		SendMessage(bot, msg.Chat.ID, "Событие не активно.")
		return
	}

	// Проверяем, участвовал ли пользователь.
	participated, err := db.UserParticipatedInEvent(eventID, msg.From.ID)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка проверки участия: "+err.Error())
		return
	}
	if participated {
		SendMessage(bot, msg.Chat.ID, "Вы уже приняли участие в этом событии.")
		return
	}

	// Получаем профиль пользователя.
	profile, err := db.GetProfile(msg.From.ID)
	if err != nil || profile == nil {
		SendMessage(bot, msg.Chat.ID, "Профиль не найден. Зарегистрируйтесь через /start.")
		return
	}

	// Начисляем валюту.
	switch event.CurrencyType {
	case "piastres", "пиastres":
		profile.Piastres += event.Amount
	case "oblomki", "обломки":
		profile.Oblomki += event.Amount
	default:
		SendMessage(bot, msg.Chat.ID, "Неизвестный тип валюты в событии.")
		return
	}

	// Сохраняем профиль и участие.
	err = db.SaveProfile(profile)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка сохранения профиля: "+err.Error())
		return
	}

	err = db.AddEventParticipation(eventID, msg.From.ID)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка регистрации участия: "+err.Error())
		return
	}

	SendMessage(bot, msg.Chat.ID, "Вы успешно приняли участие в событии!\nВаш профиль:\n"+utils.FormatProfile(profile))
}

// HandleUnattendEvent обрабатывает команду /unattend <event_id>.
func HandleUnattendEvent(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	arg := strings.TrimSpace(msg.CommandArguments())
	if arg == "" {
		SendMessage(bot, msg.Chat.ID, "Используйте: /unattend <event_id>")
		return
	}
	eventID, err := strconv.Atoi(arg)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Event ID должно быть числом.")
		return
	}
	event, err := db.GetEventByID(eventID)
	if err != nil || event == nil {
		SendMessage(bot, msg.Chat.ID, "Событие не найдено.")
		return
	}
	if !event.Active {
		SendMessage(bot, msg.Chat.ID, "Событие не активно.")
		return
	}

	// Проверяем, участвовал ли пользователь.
	participated, err := db.UserParticipatedInEvent(eventID, msg.From.ID)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка проверки участия: "+err.Error())
		return
	}
	if !participated {
		SendMessage(bot, msg.Chat.ID, "Вы не принимали участие в этом событии.")
		return
	}

	// Получаем профиль пользователя.
	profile, err := db.GetProfile(msg.From.ID)
	if err != nil || profile == nil {
		SendMessage(bot, msg.Chat.ID, "Профиль не найден.")
		return
	}

	// Списываем валюту.

	// Сохраняем изменения.
	err = db.SaveProfile(profile)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка сохранения профиля: "+err.Error())
		return
	}

	err = db.RemoveEventParticipation(eventID, msg.From.ID)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка отмены участия: "+err.Error())
		return
	}

	SendMessage(bot, msg.Chat.ID, "Вы отменили участие в событии. Валюта списана.\nВаш профиль:\n"+utils.FormatProfile(profile))
}
