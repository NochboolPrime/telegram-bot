package main

import (
	"log"
	"telegram-bot/config"
	"telegram-bot/db"
	"telegram-bot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Инициализация основного бота
	botToken := config.LoadConfig()
	if botToken == "" {
		botToken = "7785304405:AAFWCYVT_s8axaV8P4exA54TrkvB_Q5rRco"
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Основной бот авторизован как: %s", bot.Self.UserName)

	// Инициализация второго админского бота с новым API‑ключом
	secondBotToken := "8051322387:AAG4pnS8hch0JHBWgVS1qLt12JQCjd_JyB0"
	secondBot, err := tgbotapi.NewBotAPI(secondBotToken)
	if err != nil {
		log.Panic(err)
	}
	secondBot.Debug = true
	log.Printf("Админский бот авторизован как: %s", secondBot.Self.UserName)

	// Передаём и основной, и второй ботов в пакет handlers.
	// Для второго бота обязательно присваиваем глобальную переменную SecondBot из handlers.
	handlers.SecondBot = secondBot

	// Инициализация базы данных
	db.InitDB()

	// Запускаем получение обновлений основного бота
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			go handlers.HandleUpdate(bot, update)
		}
	}()

	// Запускаем отдельный цикл обновлений для второго админского бота
	u2 := tgbotapi.NewUpdate(0)
	u2.Timeout = 60
	adminUpdates := secondBot.GetUpdatesChan(u2)
	go func() {
		for update := range adminUpdates {
			go handlers.HandleAdminUpdate(secondBot, update)
		}
	}()

	// Блокировка main-потока (например, через select{}, чтобы программа не завершалась)
	select {}
}
