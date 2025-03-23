package models

import "time"

// Event описывает событие начисления валюты.
type Event struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	CurrencyType string    `json:"currency_type"` // Например, "piastres" или "oblomki"
	Amount       int       `json:"amount"`        // Сумма валюты, которую надо начислить
	Active       bool      `json:"active"`        // Флаг активности события
	CreatedAt    time.Time `json:"created_at"`    // Дата создания события
}
