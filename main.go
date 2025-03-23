package main

import (
	"log"
	"telegram-bot/config"
	"telegram-bot/db"
	"telegram-bot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Инициализация пользовательского (основного) бота
	botToken := config.LoadConfig()
	if botToken == "" {
		botToken = "7785304405:AAFWCYVT_s8axaV8P4exA54TrkvB_Q5rRco"
	}
	primaryBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	primaryBot.Debug = true
	log.Printf("Пользовательский бот авторизован как: %s", primaryBot.Self.UserName)
	handlers.PrimaryBot = primaryBot

	// Инициализация админского бота
	adminBotToken := "8051322387:AAG4pnS8hch0JHBWgVS1qLt12JQCjd_JyB0"
	adminBot, err := tgbotapi.NewBotAPI(adminBotToken)
	if err != nil {
		log.Panic(err)
	}
	adminBot.Debug = true
	log.Printf("Админский бот авторизован как: %s", adminBot.Self.UserName)
	handlers.AdminBot = adminBot

	// Инициализация базы данных
	db.InitDB()

	// Запускаем цикл обновлений для пользовательского бота
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := primaryBot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			// Обработка обновлений для пользовательского бота
			go handlers.HandleUpdate(primaryBot, update)
		}
	}()

	// Запускаем цикл обновлений для админского бота
	u2 := tgbotapi.NewUpdate(0)
	u2.Timeout = 60
	adminUpdates := adminBot.GetUpdatesChan(u2)
	go func() {
		for update := range adminUpdates {
			// Обработка обновлений для админского бота
			go handlers.HandleAdminUpdate(adminBot, update)
		}
	}()

	// Блокировка главного потока
	select {}
}
