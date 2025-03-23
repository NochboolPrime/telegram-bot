package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"telegram-bot/db"
	"telegram-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCreateEvent создает новое событие и рассылает уведомление всем пользователям.
// Формат команды (админская команда):
//
//	/createevent Название события|валюта|количество
func HandleCreateEvent(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	parts := strings.Split(args, "|")
	if len(parts) != 3 {
		SendMessage(bot, msg.Chat.ID, "Используйте формат: /createevent Название события|валюта|количество")
		return
	}

	name := strings.TrimSpace(parts[0])
	currency := strings.ToLower(strings.TrimSpace(parts[1]))
	amount, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Количество должно быть числом.")
		return
	}

	// Создаем объект события
	event := &models.Event{
		Name:         name,
		CurrencyType: currency,
		Amount:       amount,
		Active:       true,
		CreatedAt:    time.Now(),
	}

	err = db.CreateEvent(event)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка создания события: "+err.Error())
		return
	}

	log.Printf("Создано событие: ID=%d, Название=%s, Валюта=%s, Сумма=%d", event.ID, event.Name, event.CurrencyType, event.Amount)

	// Формируем текст уведомления для пользователей.
	notifyText := fmt.Sprintf("Новое событие: \"%s\" (ID: %d)\nДля участия введите: /attend %d\nДля отмены участия: /unattend %d",
		event.Name, event.ID, event.ID, event.ID)

	// Рассылка уведомления всем зарегистрированным пользователям.
	profiles, err := db.GetAllProfiles()
	if err != nil || len(profiles) == 0 {
		SendMessage(bot, msg.Chat.ID, "Ошибка рассылки уведомления или нет зарегистрированных пользователей.")
		return
	}

	for _, profile := range profiles {
		notification := tgbotapi.NewMessage(profile.TelegramID, notifyText)
		bot.Send(notification)
	}

}
