package handlers

import (
	"log"
	"strconv"
	"telegram-bot/db"
	"telegram-bot/models"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var SecondBot *tgbotapi.BotAPI

const (
	StepName = iota
	StepAge
	StepHeight
	StepWeight
	StepInventory
	StepPhoto
	StepRank
	StepTeam
	StepCompleted
)

type ConversationState struct {
	CurrentStep int
	Profile     *models.Profile
}

var registrationStates = make(map[int64]*ConversationState)

// HandleStart – при команде /start выводится меню с командами и начинается регистрация (если анкета отсутствует)
func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	helpText := "Доступные команды:\n" +
		"/start - показать меню и начать регистрацию\n" +
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
		"/help - вывести это сообщение\n"

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

// HandleUserHelp – обрабатывает команду /help для пользователского бота.
func HandleUserHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	helpText := "Доступные команды:\n" +
		"/start - показать меню и начать регистрацию\n" +
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
		"/help - вывести это сообщение"
	SendMessage(bot, msg.Chat.ID, helpText)
}

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
		state.CurrentStep = StepCompleted

		err := db.SaveProfile(state.Profile)
		if err != nil {
			log.Printf("Ошибка сохранения профиля для telegram_id %d: %v", msg.From.ID, err)
			SendMessage(bot, chatID, "Ошибка сохранения профиля. Попробуйте снова. (Ошибка: "+err.Error()+")")
		} else {
			SendMessage(bot, chatID, "Профиль успешно создан!")
			// Для пользователя используем форматирование без ID и TG username.
			userProfileText := utils.FormatProfile(state.Profile)
			SendMessage(bot, chatID, userProfileText)
			// Полная информация (для администратора) отправляется второму боту.
			SendProfileToSecondBot(state.Profile)
		}
		delete(registrationStates, msg.From.ID)
	}
}

func HandleDeleteProfile(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	err := db.DeleteProfile(msg.From.ID)
	if err != nil {
		SendMessage(bot, msg.Chat.ID, "Ошибка при удалении профиля: "+err.Error())
	} else {
		SendMessage(bot, msg.Chat.ID, "Профиль успешно удалён.")
	}
}

func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	message := tgbotapi.NewMessage(chatID, text)
	bot.Send(message)
}

// Отправка анкеты второму админскому боту (с полной информацией, включая ID и TG).
func SendProfileToSecondBot(profile *models.Profile) {
	if SecondBot == nil {
		log.Println("Второй админский бот не инициализирован")
		return
	}
	profileText := utils.FormatProfileAdmin(profile)
	// Укажите фактический chat_id для получения сообщений администратором второго бота.
	secondAdminChatID := int64(123456789) // Замените на реальный chat_id
	if profile.Photo != "" {
		photoMsg := tgbotapi.NewPhoto(secondAdminChatID, tgbotapi.FileID(profile.Photo))
		photoMsg.Caption = profileText
		_, err := SecondBot.Send(photoMsg)
		if err != nil {
			log.Printf("Ошибка отправки фото профиля второму боту: %v", err)
		}
	} else {
		message := tgbotapi.NewMessage(secondAdminChatID, profileText)
		_, err := SecondBot.Send(message)
		if err != nil {
			log.Printf("Ошибка отправки профиля второму боту: %v", err)
		}
	}
}
