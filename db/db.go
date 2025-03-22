package db

import (
	"database/sql"
	"errors"
	"log"
	"telegram-bot/models"

	_ "modernc.org/sqlite" // чисто-Go драйвер SQLite
)

var DB *sql.DB

// InitDB открывает (или создаёт) базу данных и таблицу profiles.
func InitDB() {
	var err error
	// Используйте DSN с нужными параметрами, чтобы база создавалась в режиме чтения/записи.
	DB, err = sql.Open("sqlite", "file:bot.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}

	query := `
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
        piastres INTEGER,
        oblomki INTEGER,
        attendance_count INTEGER
    );`

	_, err = DB.Exec(query)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы: %v", err)
	}
}

// GetProfile извлекает профиль по telegram_id.
func GetProfile(telegramID int64) (*models.Profile, error) {
	query := `
    SELECT id, telegram_id, username, name, age, height, weight, inventory, photo, rank, team, piastres, oblomki, attendance_count 
    FROM profiles WHERE telegram_id = ?`
	row := DB.QueryRow(query, telegramID)

	var p models.Profile
	err := row.Scan(&p.ID, &p.TelegramID, &p.Username, &p.Name, &p.Age, &p.Height, &p.Weight,
		&p.Inventory, &p.Photo, &p.Rank, &p.Team, &p.Piastres, &p.Oblomki, &p.AttendanceCount)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProfileByID извлекает профиль по уникальному номеру (ID).
func GetProfileByID(id int) (*models.Profile, error) {
	query := `
    SELECT id, telegram_id, username, name, age, height, weight, inventory, photo, rank, team, piastres, oblomki, attendance_count
    FROM profiles WHERE id = ?`
	row := DB.QueryRow(query, id)

	var p models.Profile
	err := row.Scan(&p.ID, &p.TelegramID, &p.Username, &p.Name, &p.Age, &p.Height, &p.Weight,
		&p.Inventory, &p.Photo, &p.Rank, &p.Team, &p.Piastres, &p.Oblomki, &p.AttendanceCount)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProfile вставляет новый профиль в базу.
func CreateProfile(p *models.Profile) error {
	query := `
    INSERT INTO profiles (telegram_id, username, name, age, height, weight, inventory, photo, rank, team, piastres, oblomki, attendance_count)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, p.TelegramID, p.Username, p.Name, p.Age, p.Height, p.Weight,
		p.Inventory, p.Photo, p.Rank, p.Team, p.Piastres, p.Oblomki, p.AttendanceCount)
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
    UPDATE profiles SET username = ?, name = ?, age = ?, height = ?, weight = ?, inventory = ?, photo = ?, rank = ?, team = ?, piastres = ?, oblomki = ?, attendance_count = ?
    WHERE telegram_id = ?`
	_, err := DB.Exec(query, p.Username, p.Name, p.Age, p.Height, p.Weight, p.Inventory,
		p.Photo, p.Rank, p.Team, p.Piastres, p.Oblomki, p.AttendanceCount, p.TelegramID)
	return err
}

// SaveProfile сохраняет профиль: создаёт новый, если профиль не найден, или обновляет существующий.
func SaveProfile(p *models.Profile) error {
	_, err := GetProfile(p.TelegramID)
	if err != nil {
		// Если ошибка sql.ErrNoRows, значит профиль отсутствует.
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
    SELECT id, telegram_id, username, name, age, height, weight, inventory, photo, rank, team, piastres, oblomki, attendance_count
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
			&p.Inventory, &p.Photo, &p.Rank, &p.Team, &p.Piastres, &p.Oblomki, &p.AttendanceCount)
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
