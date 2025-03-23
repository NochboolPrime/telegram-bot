package db

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"telegram-bot/models"

	_ "modernc.org/sqlite" // чисто-Go драйвер SQLite
)

var DB *sql.DB

// InitDB открывает (или создаёт) базу данных и создает таблицы.
func InitDB() {
	var err error
	// Используйте DSN с нужными параметрами, чтобы база создавалась в режиме чтения/записи.
	DB, err = sql.Open("sqlite", "file:bot.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}

	// Таблица для профилей пользователей.
	queryProfiles := `
CREATE TABLE IF NOT EXISTS profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER UNIQUE,
    username TEXT,
    name TEXT,
    age INTEGER,
    height REAL,
    weight REAL,
    inventory TEXT,
    photo TEXT,
    rank TEXT,
    team TEXT,
    race TEXT,
    piastres INTEGER,
    oblomki INTEGER
);`
	_, err = DB.Exec(queryProfiles)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы profiles: %v", err)
	}

	// Таблица для событий.
	queryEvents := `
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    currency_type TEXT,
    amount INTEGER,
    active INTEGER,
    created_at DATETIME
);`
	_, err = DB.Exec(queryEvents)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы events: %v", err)
	}

	// Таблица для участия в событиях.
	queryParticipation := `
CREATE TABLE IF NOT EXISTS event_participation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER,
    telegram_id INTEGER,
    UNIQUE(event_id, telegram_id)
);`
	_, err = DB.Exec(queryParticipation)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы event_participation: %v", err)
	}
}

// -------------------- Функции для работы с профилями --------------------------

// GetProfile извлекает профиль по telegram_id.
func GetProfile(telegramID int64) (*models.Profile, error) {
	query := `
    SELECT id, telegram_id, username, name, age, height, weight, inventory, photo, rank, team, race, piastres, oblomki
    FROM profiles WHERE telegram_id = ?`
	row := DB.QueryRow(query, telegramID)

	var p models.Profile
	err := row.Scan(&p.ID, &p.TelegramID, &p.Username, &p.Name, &p.Age, &p.Height, &p.Weight,
		&p.Inventory, &p.Photo, &p.Rank, &p.Team, &p.Race, &p.Piastres, &p.Oblomki)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProfileByID извлекает профиль по уникальному номеру (ID).
func GetProfileByID(id int) (*models.Profile, error) {
	query := `
    SELECT id, telegram_id, username, name, age, height, weight, inventory, photo, rank, team, race, piastres, oblomki
    FROM profiles WHERE id = ?`
	row := DB.QueryRow(query, id)

	var p models.Profile
	err := row.Scan(&p.ID, &p.TelegramID, &p.Username, &p.Name, &p.Age, &p.Height, &p.Weight,
		&p.Inventory, &p.Photo, &p.Rank, &p.Team, &p.Race, &p.Piastres, &p.Oblomki)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProfile вставляет новый профиль в базу.
func CreateProfile(p *models.Profile) error {
	query := `
    INSERT INTO profiles (telegram_id, username, name, age, height, weight, inventory, photo, rank, team, race, piastres, oblomki)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, p.TelegramID, p.Username, p.Name, p.Age, p.Height, p.Weight,
		p.Inventory, p.Photo, p.Rank, p.Team, p.Race, p.Piastres, p.Oblomki)
	if err != nil {
		log.Printf("Ошибка вставки профиля: %v", err)
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = int(id)
	return nil
}

// UpdateProfile обновляет данные профиля.
func UpdateProfile(p *models.Profile) error {
	query := `
    UPDATE profiles SET username = ?, name = ?, age = ?, height = ?, weight = ?,
    inventory = ?, photo = ?, rank = ?, team = ?, race = ?, piastres = ?, oblomki = ?
    WHERE telegram_id = ?`
	_, err := DB.Exec(query, p.Username, p.Name, p.Age, p.Height, p.Weight, p.Inventory,
		p.Photo, p.Rank, p.Team, p.Race, p.Piastres, p.Oblomki, p.TelegramID)
	return err
}

// SaveProfile сохраняет профиль: создает новый, если профиль не найден, или обновляет существующий.
func SaveProfile(p *models.Profile) error {
	_, err := GetProfile(p.TelegramID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CreateProfile(p)
		}
		return err
	}
	return UpdateProfile(p)
}

// GetAllProfiles возвращает все профили.
func GetAllProfiles() ([]*models.Profile, error) {
	query := `
    SELECT id, telegram_id, username, name, age, height, weight, inventory, photo, rank, team, race, piastres, oblomki
    FROM profiles`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*models.Profile
	for rows.Next() {
		var p models.Profile
		err = rows.Scan(&p.ID, &p.TelegramID, &p.Username, &p.Name, &p.Age, &p.Height, &p.Weight,
			&p.Inventory, &p.Photo, &p.Rank, &p.Team, &p.Race, &p.Piastres, &p.Oblomki)
		if err != nil {
			log.Printf("Ошибка Scan: %v", err)
			continue
		}
		profiles = append(profiles, &p)
	}
	return profiles, nil
}

// DeleteProfile удаляет профиль по telegram_id.
func DeleteProfile(telegramID int64) error {
	query := "DELETE FROM profiles WHERE telegram_id = ?"
	res, err := DB.Exec(query, telegramID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return errors.New("профиль не найден")
	}
	return nil
}

// DeleteProfileByID удаляет профиль по уникальному номеру (id).
func DeleteProfileByID(id int) error {
	query := "DELETE FROM profiles WHERE id = ?"
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("профиль не найден")
	}
	return nil
}

// -------------------- Функции для работы с событиями --------------------------

// boolToInt преобразует bool в int (1 - true, 0 - false).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// CreateEvent вставляет новое событие в базу.
func CreateEvent(e *models.Event) error {
	query := `
INSERT INTO events (name, currency_type, amount, active, created_at)
VALUES (?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, e.Name, e.CurrencyType, e.Amount, boolToInt(e.Active), e.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.ID = int(id)
	return nil
}

// GetEventByID извлекает событие из базы по его ID.
func GetEventByID(id int) (*models.Event, error) {
	query := `
SELECT id, name, currency_type, amount, active, created_at
FROM events WHERE id = ?`
	row := DB.QueryRow(query, id)
	var e models.Event
	var activeInt int
	var createdAtStr string
	err := row.Scan(&e.ID, &e.Name, &e.CurrencyType, &e.Amount, &activeInt, &createdAtStr)
	if err != nil {
		return nil, err
	}
	e.Active = activeInt == 1
	e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEvent обновляет событие (например, завершает его).
func UpdateEvent(e *models.Event) error {
	query := `
UPDATE events SET name = ?, currency_type = ?, amount = ?, active = ?, created_at = ?
WHERE id = ?`
	_, err := DB.Exec(query, e.Name, e.CurrencyType, e.Amount, boolToInt(e.Active), e.CreatedAt.Format(time.RFC3339), e.ID)
	return err
}

// GetActiveEvents возвращает список активных событий.
func GetActiveEvents() ([]*models.Event, error) {
	query := `
SELECT id, name, currency_type, amount, active, created_at
FROM events WHERE active = 1`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []*models.Event
	for rows.Next() {
		var e models.Event
		var activeInt int
		var createdAtStr string
		err = rows.Scan(&e.ID, &e.Name, &e.CurrencyType, &e.Amount, &activeInt, &createdAtStr)
		if err != nil {
			log.Printf("Ошибка сканирования события: %v", err)
			continue
		}
		e.Active = activeInt == 1
		e.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Printf("Ошибка парсинга даты: %v", err)
			continue
		}
		events = append(events, &e)
	}
	return events, nil
}

// UserParticipatedInEvent проверяет, отмечался ли пользователь (telegram_id) на событие (event_id).
func UserParticipatedInEvent(eventID int, telegramID int64) (bool, error) {
	query := `
SELECT COUNT(*) FROM event_participation 
WHERE event_id = ? AND telegram_id = ?`
	row := DB.QueryRow(query, eventID, telegramID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddEventParticipation регистрирует участие пользователя в событии.
func AddEventParticipation(eventID int, telegramID int64) error {
	query := `
INSERT INTO event_participation (event_id, telegram_id)
VALUES (?, ?)`
	_, err := DB.Exec(query, eventID, telegramID)
	return err
}

// RemoveEventParticipation удаляет запись о участии пользователя в событии.
func RemoveEventParticipation(eventID int, telegramID int64) error {
	query := `
DELETE FROM event_participation 
WHERE event_id = ? AND telegram_id = ?`
	res, err := DB.Exec(query, eventID, telegramID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("участие не найдено")
	}
	return nil
}
