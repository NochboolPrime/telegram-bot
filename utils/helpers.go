package utils

import (
	"fmt"
	"telegram-bot/models"
)

// FormatProfile – форматирует анкету для общего просмотра (без уникального номера и username)
func FormatProfile(p *models.Profile) string {
	return fmt.Sprintf("Анкета персонажа:\n"+
		"Имя: %s\n"+
		"Возраст: %d\n"+
		"Рост: %.2f\n"+
		"Вес: %.2f\n"+
		"Инвентарь: %s\n"+
		"Ранг: %s\n"+
		"Команда: %s\n"+
		"Раса: %s\n"+
		"Пиастры: %d\n"+
		"Обломки: %d",
		p.Name, p.Age, p.Height, p.Weight,
		p.Inventory, p.Rank, p.Team, p.Race, p.Piastres, p.Oblomki)
}

// FormatProfileAdmin – форматирует анкету для администратора,
// добавляя уникальный номер (ID) и TG username.
func FormatProfileAdmin(p *models.Profile) string {
	return fmt.Sprintf("Анкета персонажа (ID: %d, TG: @%s):\n"+
		"Имя: %s\n"+
		"Возраст: %d\n"+
		"Рост: %.2f\n"+
		"Вес: %.2f\n"+
		"Инвентарь: %s\n"+
		"Ранг: %s\n"+
		"Команда: %s\n"+
		"Раса: %s\n"+
		"Пиастры: %d\n"+
		"Обломки: %d",
		p.ID, p.Username, p.Name, p.Age, p.Height, p.Weight,
		p.Inventory, p.Rank, p.Team, p.Race, p.Piastres, p.Oblomki)
}
