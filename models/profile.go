package models

// Profile описывает анкету персонажа.
// Поле ID – уникальный номер анкеты (из базы),
// Username – имя пользователя в Telegram (видно только администраторам).
type Profile struct {
	ID         int
	TelegramID int64
	Username   string
	Name       string
	Age        int
	Height     float64
	Weight     float64
	Inventory  string
	Photo      string // file_id или URL фотографии
	Rank       string
	Team       string
	Race       string // Новое поле: раса
	Piastres   int
	Oblomki    int
}
