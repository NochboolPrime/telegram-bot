package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"telegram-bot/db"
	"telegram-bot/models"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const AdminPassword = "SuperSecret123"

// Глобальная карта для хранения аутентифицированных админов.
var authenticatedAdmins = make(map[int64]bool)

// isAuthenticated проверяет, аутентифицирован ли данный chatID.
func isAuthenticated(chatID int64) bool {
	return authenticatedAdmins[chatID]
}

// HandleCreateEvent создает новое событие в базе данных.
// Ожидаемый формат команды: /createevent <название|валюта|количество>
func HandleCreateEvent(bot *tgbotapi.BotAPI, chatID int64, args string) {
	parts := strings.Split(args, "|")
	if len(parts) != 3 {
		msg := tgbotapi.NewMessage(chatID, "Используйте формат команды: /createevent <название|валюта|количество>")
		bot.Send(msg)
		return
	}

	name := strings.TrimSpace(parts[0])
	currency := strings.ToLower(strings.TrimSpace(parts[1]))
	amount, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Количество должно быть числом.")
		bot.Send(msg)
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

	// Сохраняем событие в базе данных
	err = db.CreateEvent(event)
	if err != nil {
		log.Printf("Ошибка создания события: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка создания события: "+err.Error())
		bot.Send(msg)
		return
	}

	log.Printf("Создано событие: ID=%d, Название=%s, Валюта=%s, Сумма=%d", event.ID, event.Name, event.CurrencyType, event.Amount)

	// Уведомление для админа
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Событие создано: \"%s\" (ID: %d)", event.Name, event.ID))
	bot.Send(msg)

	// Отправляем уведомление всем пользователям
	notifyText := fmt.Sprintf("Новое событие: \"%s\" (ID: %d)\n"+
		"Для участия введите: /attend %d\n"+
		"Для отказа: /unattend %d", event.Name, event.ID, event.ID, event.ID)

	profiles, err := db.GetAllProfiles()
	if err != nil {
		log.Printf("Ошибка получения профилей для рассылки: %v", err)
		return
	}

	for _, profile := range profiles {
		notification := tgbotapi.NewMessage(profile.TelegramID, notifyText)
		bot.Send(notification)
	}
}

// authenticate выполняет аутентификацию: если предоставленный пароль совпадает с AdminPassword,
// то chatID сохраняется как аутентифицированный.
func authenticate(chatID int64, providedPassword string) bool {
	if providedPassword == AdminPassword {
		authenticatedAdmins[chatID] = true
		return true
	}
	return false
}

// HandleAdminUpdate обрабатывает команды администраторского бота.
func HandleAdminUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	chatID := update.Message.Chat.ID
	cmd := strings.ToLower(update.Message.Command())
	args := update.Message.CommandArguments()

	// Обработка команды /help всегда доступна.
	if cmd == "help" {
		helpText := "Доступные команды админского бота:\n" +
			"/auth <пароль> - аутентификация\n" +
			"/help - показать это сообщение\n" +
			"/allprofiles - показать список всех анкет (ID, TG username)\n" +
			"/viewprofile <ID> - просмотр анкеты по ID\n" +
			"/editprofile <ID> <поле> <значение> - редактирование анкеты\n" +
			"/deleteprofilebyid <ID> - удаление анкеты по ID\n" +
			"/addcurrency <ID> <тип валюты> <количество> - добавление валюты\n" +
			"/createevent <название|валюта|сумма> - создание события\n"
		msg := tgbotapi.NewMessage(chatID, helpText)
		bot.Send(msg)
		return
	}

	// Команда /auth всегда доступна.
	if cmd == "auth" {
		password := strings.TrimSpace(args)
		if authenticate(chatID, password) {
			msg := tgbotapi.NewMessage(chatID, "Аутентификация успешна!")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "Неверный пароль. Попробуйте снова: /auth <пароль>")
			bot.Send(msg)
		}
		return
	}

	// Если админ не аутентифицирован, остальные команды недоступны.
	if !isAuthenticated(chatID) {
		msg := tgbotapi.NewMessage(chatID, "Для доступа к функциям админского бота необходимо аутентифицироваться. Введите: /auth <пароль>")
		bot.Send(msg)
		return
	}

	// Обработка других команд
	switch cmd {
	case "allprofiles":
		profiles, err := db.GetAllProfiles()
		if err != nil {
			log.Printf("Ошибка получения профилей: %v", err)
			msg := tgbotapi.NewMessage(chatID, "Ошибка получения профилей.")
			bot.Send(msg)
			return
		}
		if len(profiles) == 0 {
			msg := tgbotapi.NewMessage(chatID, "Нет созданных анкет.")
			bot.Send(msg)
			return
		}
		var listText string
		for _, p := range profiles {
			listText += "ID: " + strconv.Itoa(p.ID) + " | TG: @" + p.Username + "\n"
		}
		prompt := "Для просмотра анкеты отправьте команду: /viewprofile <ID>\n" +
			"Для редактирования: /editprofile <ID> <поле> <значение>\n" +
			"Для удаления: /deleteprofilebyid <ID>\n" +
			"Для добавления валюты: /addcurrency <ID> <тип валюты> <количество>\n"
		fullText := listText + "\n" + prompt
		msg := tgbotapi.NewMessage(chatID, fullText)
		bot.Send(msg)

	case "viewprofile":
		parts := strings.Fields(args)
		if len(parts) != 1 {
			msg := tgbotapi.NewMessage(chatID, "Используйте: /viewprofile <ID>")
			bot.Send(msg)
			return
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный ID анкеты.")
			bot.Send(msg)
			return
		}
		profile, err := db.GetProfileByID(id)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Ошибка получения анкеты: "+err.Error())
			bot.Send(msg)
			return
		}
		detailText := utils.FormatProfileAdmin(profile)
		msg := tgbotapi.NewMessage(chatID, detailText)
		bot.Send(msg)

	case "deleteprofilebyid":
		parts := strings.Fields(args)
		if len(parts) != 1 {
			msg := tgbotapi.NewMessage(chatID, "Используйте: /deleteprofilebyid <ID>")
			bot.Send(msg)
			return
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный ID анкеты.")
			bot.Send(msg)
			return
		}
		err = db.DeleteProfileByID(id)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Ошибка удаления анкеты: "+err.Error())
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "Анкета с ID "+strconv.Itoa(id)+" успешно удалена.")
			bot.Send(msg)
		}

	case "editprofile":
		parts := strings.Fields(args)
		if len(parts) < 3 {
			msg := tgbotapi.NewMessage(chatID, "Используйте: /editprofile <ID> <поле> <значение>")
			bot.Send(msg)
			return
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный ID анкеты.")
			bot.Send(msg)
			return
		}
		field := strings.ToLower(parts[1])
		newValue := strings.Join(parts[2:], " ")

		profile, err := db.GetProfileByID(id)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Ошибка получения анкеты: "+err.Error())
			bot.Send(msg)
			return
		}

		edited := false
		switch field {
		case "name":
			profile.Name = newValue
			edited = true
		case "age":
			age, err := strconv.Atoi(newValue)
			if err == nil {
				profile.Age = age
				edited = true
			}
		case "height":
			height, err := strconv.ParseFloat(newValue, 64)
			if err == nil {
				profile.Height = height
				edited = true
			}
		case "weight":
			weight, err := strconv.ParseFloat(newValue, 64)
			if err == nil {
				profile.Weight = weight
				edited = true
			}
		case "inventory":
			profile.Inventory = newValue
			edited = true
		case "photo":
			profile.Photo = newValue
			edited = true
		case "rank":
			profile.Rank = newValue
			edited = true
		case "team":
			profile.Team = newValue
			edited = true
		default:
			msg := tgbotapi.NewMessage(chatID, "Неизвестное поле. Доступны: name, age, height, weight, inventory, photo, rank, team")
			bot.Send(msg)
			return
		}
		if edited {
			err = db.UpdateProfile(profile)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Ошибка обновления анкеты: "+err.Error())
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Анкета обновлена.")
				bot.Send(msg)
			}
		} else {
			msg := tgbotapi.NewMessage(chatID, "Не удалось обновить анкету. Проверьте входные данные.")
			bot.Send(msg)
		}

	case "addcurrency":
		parts := strings.Fields(args)
		if len(parts) != 3 {
			msg := tgbotapi.NewMessage(chatID, "Используйте корректный формат: /addcurrency <ID> <тип валюты> <количество>")
			bot.Send(msg)
			return
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Неверный ID анкеты.")
			bot.Send(msg)
			return
		}
		currency := strings.ToLower(parts[1])
		amount, err := strconv.Atoi(parts[2])
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Количество должно быть числом.")
			bot.Send(msg)
			return
		}

		profile, err := db.GetProfileByID(id)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Ошибка получения анкеты: "+err.Error())
			bot.Send(msg)
			return
		}

		switch currency {
		case "piastres", "пиастры":
			profile.Piastres += amount
		case "oblomki", "обломки":
			profile.Oblomki += amount
		default:
			msg := tgbotapi.NewMessage(chatID, "Неизвестный тип валюты. Используйте 'piastres' или 'oblomki'.")
			bot.Send(msg)
			return
		}

		err = db.UpdateProfile(profile)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Ошибка обновления анкеты: "+err.Error())
			bot.Send(msg)
			return
		}
		responseText := "Валюта успешно начислена.\nНовая сумма:\nПиастры: " + strconv.Itoa(profile.Piastres) +
			"\nОбломки: " + strconv.Itoa(profile.Oblomki)
		msg := tgbotapi.NewMessage(chatID, responseText)
		bot.Send(msg)

	case "createevent":
		parts := strings.Split(args, "|")
		if len(parts) != 3 {
			msg := tgbotapi.NewMessage(chatID, "Используйте формат команды: /createevent <название|валюта|количество>")
			bot.Send(msg)
			return
		}

		name := strings.TrimSpace(parts[0])
		currency := strings.ToLower(strings.TrimSpace(parts[1]))
		amount, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "Количество должно быть числом.")
			bot.Send(msg)
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

		// Сохраняем событие в базе данных
		err = db.CreateEvent(event)
		if err != nil {
			log.Printf("Ошибка создания события: %v", err)
			msg := tgbotapi.NewMessage(chatID, "Ошибка создания события: "+err.Error())
			bot.Send(msg)
			return
		}

		log.Printf("Создано событие: ID=%d, Название=%s, Валюта=%s, Сумма=%d", event.ID, event.Name, event.CurrencyType, event.Amount)

		// Уведомление для админа
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Событие создано: \"%s\" (ID: %d)", event.Name, event.ID))
		bot.Send(msg)

		// Отправляем уведомление всем пользователям
		notifyText := fmt.Sprintf("Новое событие: \"%s\" (ID: %d)\n"+
			"Для участия введите: /attend %d\n"+
			"Для отказа: /unattend %d", event.Name, event.ID, event.ID, event.ID)

		profiles, err := db.GetAllProfiles()
		if err != nil {
			log.Printf("Ошибка получения профилей для рассылки: %v", err)
			return
		}

		for _, profile := range profiles {
			notification := tgbotapi.NewMessage(profile.TelegramID, notifyText)
			bot.Send(notification)
		}

	default:
		msg := tgbotapi.NewMessage(chatID, "Неизвестная команда. Используйте /help для списка доступных команд.")
		bot.Send(msg)
	}
}
