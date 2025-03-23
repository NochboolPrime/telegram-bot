package handlers

import (
	"log"
	"strconv"

	"telegram-bot/db"
	"telegram-bot/models"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Регистрационные шаги
const (
	StepName = iota
	StepAge
	StepHeight
	StepWeight
	StepInventory
	StepPhoto
	StepRank
	StepTeam
	StepRace // новый шаг: ввод расы
	StepCompleted
)

type ConversationState struct {
	CurrentStep int
	Profile     *models.Profile
}

var registrationStates = make(map[int64]*ConversationState)

// HandleStart – при команде /start запускается регистрация или выводится меню
func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	helpText := "Доступные команды:\n" +
		"/start - начать регистрацию / показать меню\n" +
		"/createprofile - создать новую анкету\n" +
		"/profile - просмотр анкеты\n" +
		"/deleteprofile - удаление анкеты\n" +
		"/setname <имя> - изменить имя\n" +
		"/setage <возраст> - изменить возраст\n" +
		"/setheight <рост> - изменить рост\n" +
		"/setweight <вес> - изменить вес\n" +
		"/setinventory <инвентарь> - изменить инвентарь\n" +
		"/setphoto <file_id> - изменить фото\n" +
		"/setrank <ранг> - изменить ранг\n" +
		"/setteam <команда> - изменить команду\n" +
		"/attend - отметиться на активном ивенте\n" +
		"/unattend - отменить отметку на активном ивенте\n" +
		"/help - вывести эту справку\n"

	existingProfile, err := db.GetProfile(msg.From.ID)
	if err == nil && existingProfile != nil && existingProfile.Name != "" {
		SendMessage(bot, msg.Chat.ID, "Привет! Ваша анкета уже создана.\n\n"+helpText)
		return
	}

	newProfile := &models.Profile{
		TelegramID: msg.From.ID,
		Username:   msg.From.UserName,
		Piastres:   0,
		Oblomki:    0,
	}
	registrationStates[msg.From.ID] = &ConversationState{
		CurrentStep: StepName,
		Profile:     newProfile,
	}
	SendMessage(bot, msg.Chat.ID, "Добро пожаловать! Начинаем регистрацию.\nВведите ваше имя.\n\n"+helpText)
}

// HandleUserHelp – выводит справку команд для пользователя.
func HandleUserHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	helpText := "Доступные команды:\n" +
		"/start - начать регистрацию / показать меню\n" +
		"/createprofile - создать новую анкету\n" +
		"/profile - просмотр анкеты\n" +
		"/deleteprofile - удаление анкеты\n" +
		"/setname <имя> - изменить имя\n" +
		"/setage <возраст> - изменить возраст\n" +
		"/setheight <рост> - изменить рост\n" +
		"/setweight <вес> - изменить вес\n" +
		"/setinventory <инвентарь> - изменить инвентарь\n" +
		"/setphoto <file_id> - изменить фото\n" +
		"/setrank <ранг> - изменить ранг\n" +
		"/setteam <команда> - изменить команду\n" +
		"/attend - отметиться на активном ивенте\n" +
		"/unattend - отменить отметку на активном ивенте\n" +
		"/help - вывести эту справку\n"
	SendMessage(bot, msg.Chat.ID, helpText)
}

// HandleRegistrationConversation обрабатывает ввод данных при регистрации.
func HandleRegistrationConversation(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, state *ConversationState) {
	chatID := msg.Chat.ID
	text := msg.Text

	switch state.CurrentStep {
	case StepName:
		state.Profile.Name = text
		state.CurrentStep = StepAge
		SendMessage(bot, chatID, "Введите ваш возраст (целое число):")
	case StepAge:
		age, err := strconv.Atoi(text)
		if err != nil {
			SendMessage(bot, chatID, "Возраст должен быть числом. Попробуйте ещё раз:")
			return
		}
		state.Profile.Age = age
		state.CurrentStep = StepHeight
		SendMessage(bot, chatID, "Введите ваш рост (например, 175.5):")
	case StepHeight:
		height, err := strconv.ParseFloat(text, 64)
		if err != nil {
			SendMessage(bot, chatID, "Рост должен быть числом. Попробуйте ещё раз:")
			return
		}
		state.Profile.Height = height
		state.CurrentStep = StepWeight
		SendMessage(bot, chatID, "Введите ваш вес (например, 70.2):")
	case StepWeight:
		weight, err := strconv.ParseFloat(text, 64)
		if err != nil {
			SendMessage(bot, chatID, "Вес должен быть числом. Попробуйте ещё раз:")
			return
		}
		state.Profile.Weight = weight
		state.CurrentStep = StepInventory
		SendMessage(bot, chatID, "Опишите ваш инвентарь:")
	case StepInventory:
		state.Profile.Inventory = text
		state.CurrentStep = StepPhoto
		SendMessage(bot, chatID, "Пришлите фотографию или введите file_id:")
	case StepPhoto:
		state.Profile.Photo = text
		state.CurrentStep = StepRank
		SendMessage(bot, chatID, "Введите ваш ранг:")
	case StepRank:
		state.Profile.Rank = text
		state.CurrentStep = StepTeam
		SendMessage(bot, chatID, "Введите вашу команду:")
	case StepTeam:
		state.Profile.Team = text
		state.CurrentStep = StepRace
		SendMessage(bot, chatID, "Введите вашу расу:")
	case StepRace:
		state.Profile.Race = text
		state.CurrentStep = StepCompleted

		err := db.SaveProfile(state.Profile)
		if err != nil {
			SendMessage(bot, chatID, "Ошибка сохранения профиля. Попробуйте снова. (Ошибка: "+err.Error()+")")
		} else {
			SendMessage(bot, chatID, "Профиль успешно создан!")
			userProfileText := utils.FormatProfile(state.Profile)
			SendMessage(bot, chatID, userProfileText)
			// Отправляем полную информацию админскому боту
			SendProfileToAdminBot(state.Profile)
		}
		delete(registrationStates, msg.From.ID)
	}
}

// HandleDeleteProfile удаляет профиль пользователя.
func HandleDeleteProfile(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	err := db.DeleteProfile(msg.From.ID)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка при удалении профиля: "+err.Error())
	} else {
		SendMessage(bot, msg.Chat.ID, "Профиль успешно удалён.")
	}
}

// HandleAttendCommand обрабатывает команду /attend – отметиться на активном ивенте.
func HandleAttendCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if currentEvent == nil {
		SendMessage(bot, msg.Chat.ID, "На данный момент активных событий нет.")
		return
	}
	userID := msg.From.ID
	if currentEvent.Participants[userID] {
		SendMessage(bot, msg.Chat.ID, "Вы уже отметились на событие.")
		return
	}

	profile, err := db.GetProfile(userID)
	if err != nil || profile == nil {
		SendMessage(bot, msg.Chat.ID, "Профиль не найден. Пожалуйста, зарегистрируйтесь через /start.")
		return
	}

	switch currentEvent.CurrencyType {
	case "piastres", "пиastres":
		profile.Piastres += currentEvent.Amount
	case "oblomki", "обломки":
		profile.Oblomki += currentEvent.Amount
	default:
		SendMessage(bot, msg.Chat.ID, "Неизвестный тип валюты в событии.")
		return
	}

	currentEvent.Participants[userID] = true
	db.SaveProfile(profile)
	SendMessage(bot, msg.Chat.ID, "Вы успешно отметились на событие! Валюта начислена.\nВаш профиль:\n"+utils.FormatProfile(profile))
}

// HandleUnattendCommand обрабатывает команду /unattend – отменить отметку на активном ивенте.
func HandleUnattendCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if currentEvent == nil {
		SendMessage(bot, msg.Chat.ID, "На данный момент активных событий нет.")
		return
	}
	userID := msg.From.ID
	if !currentEvent.Participants[userID] {
		SendMessage(bot, msg.Chat.ID, "Вы не отметились на событие, отмена невозможна.")
		return
	}

	profile, err := db.GetProfile(userID)
	if err != nil || profile == nil {
		SendMessage(bot, msg.Chat.ID, "Профиль не найден.")
		return
	}

	delete(currentEvent.Participants, userID)
	switch currentEvent.CurrencyType {
	case "piastres", "пиastres":
		if profile.Piastres >= currentEvent.Amount {
			profile.Piastres -= currentEvent.Amount
		} else {
			profile.Piastres = 0
		}
	case "oblomki", "обломки":
		if profile.Oblomki >= currentEvent.Amount {
			profile.Oblomki -= currentEvent.Amount
		} else {
			profile.Oblomki = 0
		}
	default:
		SendMessage(bot, msg.Chat.ID, "Неизвестный тип валюты в событии.")
		return
	}
	db.SaveProfile(profile)
	SendMessage(bot, msg.Chat.ID, "Ваша отметка отменена, начисленная валюта списана.\nВаш профиль:\n"+utils.FormatProfile(profile))
}

// ProcessUserCommand диспетчер пользовательских команд.
func ProcessUserCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		HandleStart(bot, msg)
	case "help":
		HandleUserHelp(bot, msg)
	case "attend":
		HandleAttendCommand(bot, msg)
	case "unattend":
		HandleUnattendCommand(bot, msg)
	// Дополнительные команды (например, profile, createprofile) можно добавить здесь.
	default:
		SendMessage(bot, msg.Chat.ID, "Неизвестная команда. Используйте /help для списка команд.")
	}
}

// SendMessage универсальная функция отправки сообщений.
func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	message := tgbotapi.NewMessage(chatID, text)
	bot.Send(message)
}

// Declare AdminBot as a global variable
var AdminBot *tgbotapi.BotAPI
var PrimaryBot *tgbotapi.BotAPI

// SendProfileToAdminBot отправляет профиль админскому боту.
func SendProfileToAdminBot(profile *models.Profile) {
	if AdminBot == nil {
		log.Println("Админский бот не инициализирован")
		return
	}
	profileText := utils.FormatProfileAdmin(profile)
	// Укажите корректный chat_id для админских уведомлений.
	adminChatID := int64(123456789) // замените на реальный chat_id
	if profile.Photo != "" {
		photoMsg := tgbotapi.NewPhoto(adminChatID, tgbotapi.FileID(profile.Photo))
		photoMsg.Caption = profileText
		_, err := AdminBot.Send(photoMsg)
		if err != nil {
			log.Printf("Ошибка отправки фото профиля админскому боту: %v", err)
		}
	} else {
		message := tgbotapi.NewMessage(adminChatID, profileText)
		_, err := AdminBot.Send(message)
		if err != nil {
			log.Printf("Ошибка отправки профиля админскому боту: %v", err)
		}
	}
}
