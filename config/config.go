package config

import (
	"os"
	"strconv"
	"strings"
)

var (
	// Токен пользовательского бота, можно переопределить через переменную окружения BOT_TOKEN.
	BotToken string = "7785304405:AAFWCYVT_s8axaV8P4exA54TrkvB_Q5rRco"
	// Список ID администраторов (определяется через переменную окружения ADMIN_IDS).
	AdminIDs []int64
	// Идентификатор чата для рассылки событий.
	// Здесь он установлен равным 7785304405.
	BroadcastChatID int64 = 7785304405
)

// LoadConfig загружает конфигурацию из переменных окружения.
// Если переменные окружения заданы, они могут переопределить значения по умолчанию.
func LoadConfig() string {
	envToken := os.Getenv("BOT_TOKEN")
	if envToken != "" {
		BotToken = envToken
	}

	// Загружаем список администраторских ID.
	adminIDsEnv := os.Getenv("ADMIN_IDS") // Пример: "12345678,98765432"
	for _, idStr := range strings.Split(adminIDsEnv, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err == nil {
			AdminIDs = append(AdminIDs, id)
		}
	}

	// Загружаем идентификатор чата для рассылки событий.
	broadcastChatIDStr := os.Getenv("BROADCAST_CHAT_ID")
	if broadcastChatIDStr != "" {
		id, err := strconv.ParseInt(broadcastChatIDStr, 10, 64)
		if err == nil {
			BroadcastChatID = id
		}
		// Если произошла ошибка, сохраняем значение по умолчанию (7785304405).
	}

	return BotToken
}

// IsAdmin возвращает true, если userID входит в список администраторов.
func IsAdmin(userID int64) bool {
	for _, id := range AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
