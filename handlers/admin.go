package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"telegram-bot/config"
	"telegram-bot/db"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура события начисления валюты
type AttendanceEvent struct {
	EventID      string
	CurrencyType string // "piastres" или "oblomki"
	Amount       int
	Participants map[int64]bool
}

var currentEvent *AttendanceEvent

// HandleModifyCurrency позволяет администратору изменять валюту у пользователя.
// Формат команды: /modifycurrency <telegram_id> <currency> <amount>
func HandleModifyCurrency(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if !config.IsAdmin(msg.From.ID) {
		sendMessage(bot, msg.Chat.ID, "Нет прав!")
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) != 3 {
		sendMessage(bot, msg.Chat.ID, "Используйте: /modifycurrency <telegram_id> <currency> <amount>")
		return
	}
	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Неверный telegram_id")
		return
	}
	currency := strings.ToLower(args[1])
	amount, err := strconv.Atoi(args[2])
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Amount должен быть числом")
		return
	}

	profile, err := db.GetProfile(userID)
	if err != nil || profile == nil {
		sendMessage(bot, msg.Chat.ID, "Профиль не найден")
		return
	}

	if currency == "piastres" || currency == "пиастры" {
		profile.Piastres = amount
	} else if currency == "oblomki" || currency == "обломки" {
		profile.Oblomki = amount
	} else {
		sendMessage(bot, msg.Chat.ID, "Неизвестный тип валюты")
		return
	}
	err = db.SaveProfile(profile)
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Ошибка сохранения профиля")
	} else {
		sendMessage(bot, msg.Chat.ID, "Валюта успешно изменена.")
	}
}

// HandleStartEvent запускает событие начисления валюты.
// Формат команды: /startevent <currency> <amount>
func HandleStartEvent(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if !config.IsAdmin(msg.From.ID) {
		sendMessage(bot, msg.Chat.ID, "Нет прав!")
		return
	}
	args := strings.Fields(msg.CommandArguments())
	if len(args) != 2 {
		sendMessage(bot, msg.Chat.ID, "Используйте: /startevent <currency> <amount>")
		return
	}
	currency := strings.ToLower(args[0])
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Amount должен быть числом")
		return
	}

	eventID := fmt.Sprintf("%d", time.Now().Unix())
	currentEvent = &AttendanceEvent{
		EventID:      eventID,
		CurrencyType: currency,
		Amount:       amount,
		Participants: make(map[int64]bool),
	}

	// Формируем inline‑клавиатуру с кнопкой «Я был»
	button := tgbotapi.NewInlineKeyboardButtonData("Я был", "attend:"+eventID)
	row := tgbotapi.NewInlineKeyboardRow(button)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	text := fmt.Sprintf("Начинается событие!\nНажмите на кнопку \"Я был\", чтобы получить %d %s.", amount, currency)
	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ReplyMarkup = keyboard
	bot.Send(msgConfig)
}

// HandleEventCallback обрабатывает нажатие на кнопку «Я был»
func HandleEventCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return
	}
	eventID := parts[1]
	if currentEvent == nil || currentEvent.EventID != eventID {
		answer := tgbotapi.NewCallback(callback.ID, "Событие больше не активно.")
		bot.Request(answer)
		return
	}

	userID := callback.From.ID
	if currentEvent.Participants[userID] {
		answer := tgbotapi.NewCallback(callback.ID, "Вы уже отметились.")
		bot.Request(answer)
		return
	}

	currentEvent.Participants[userID] = true

	profile, err := db.GetProfile(userID)
	if err != nil || profile == nil {
		answer := tgbotapi.NewCallback(callback.ID, "Профиль не найден. Зарегистрируйтесь через /start.")
		bot.Request(answer)
		return
	}

	// Начисляем валюту
	if currentEvent.CurrencyType == "piastres" || currentEvent.CurrencyType == "пиастры" {
		profile.Piastres += currentEvent.Amount
	} else if currentEvent.CurrencyType == "oblomki" || currentEvent.CurrencyType == "обломки" {
		profile.Oblomki += currentEvent.Amount
	}
	profile.AttendanceCount++
	db.SaveProfile(profile)

	answer := tgbotapi.NewCallback(callback.ID, "Вы успешно отметились!")
	bot.Request(answer)

	// Отправляем обновлённую анкету участника
	profileText := utils.FormatProfile(profile)
	sendMessage(bot, callback.Message.Chat.ID, profileText)
}

// HandleParticipants выводит список участников текущего события
func HandleParticipants(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if !config.IsAdmin(msg.From.ID) {
		sendMessage(bot, msg.Chat.ID, "Нет прав!")
		return
	}
	if currentEvent == nil {
		sendMessage(bot, msg.Chat.ID, "Нет активного события.")
		return
	}

	var result string
	for userID := range currentEvent.Participants {
		profile, err := db.GetProfile(userID)
		if err == nil && profile != nil {
			result += fmt.Sprintf("Имя: %s, TelegramID: %d\n", profile.Name, profile.TelegramID)
		}
	}
	if result == "" {
		result = "Нет участников."
	}
	sendMessage(bot, msg.Chat.ID, "Список участников:\n"+result)
}

// HandleCurrencyRanking выводит рейтинг по заданной валюте.
// Формат команды: /currencyranking <currency>
// Если параметр не указан, по умолчанию используется "piastres".
func HandleCurrencyRanking(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if !config.IsAdmin(msg.From.ID) {
		sendMessage(bot, msg.Chat.ID, "Нет прав!")
		return
	}
	args := strings.Fields(msg.CommandArguments())
	currency := "piastres"
	if len(args) > 0 {
		currency = strings.ToLower(args[0])
	}

	profiles, err := db.GetAllProfiles()
	if err != nil {
		sendMessage(bot, msg.Chat.ID, "Ошибка получения профилей")
		return
	}

	// Сортировка профилей по валюте
	sort.Slice(profiles, func(i, j int) bool {
		if currency == "piastres" || currency == "пиастры" {
			return profiles[i].Piastres > profiles[j].Piastres
		} else if currency == "oblomki" || currency == "обломки" {
			return profiles[i].Oblomki > profiles[j].Oblomki
		}
		return false
	})

	ranking := fmt.Sprintf("Рейтинг по %s:\n", currency)
	for i, p := range profiles {
		var value int
		if currency == "piastres" || currency == "пиастры" {
			value = p.Piastres
		} else {
			value = p.Oblomki
		}
		ranking += fmt.Sprintf("%d. %s – %d\n", i+1, p.Name, value)
	}
	sendMessage(bot, msg.Chat.ID, ranking)
}
