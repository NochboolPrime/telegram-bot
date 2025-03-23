package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"telegram-bot/config"
	"telegram-bot/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

// Вспомогательная функция для отправки сообщения.
// Поскольку файлы находятся в одном пакете, функция доступна и в event.go.
