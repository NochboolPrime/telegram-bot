package config

import (
	"os"
	"strconv"
	"strings"
)

var (
	BotToken string
	AdminIDs []int64
)

// LoadConfig загружает конфигурацию из переменных окружения.
// Например, задайте переменные BOT_TOKEN и ADMIN_IDS (через запятую).
func LoadConfig() string {
	BotToken = os.Getenv("BOT_TOKEN")
	adminIDsEnv := os.Getenv("ADMIN_IDS") // Пример: "12345678,98765432"
	for _, idStr := range strings.Split(adminIDsEnv, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err == nil {
			AdminIDs = append(AdminIDs, id)
		}
	}
	return BotToken
}

func IsAdmin(userID int64) bool {
	for _, id := range AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
