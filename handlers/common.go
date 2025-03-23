package handlers

// AttendanceEvent описывает событие начисления валюты.
type AttendanceEvent struct {
	EventID      string
	CurrencyType string
	Amount       int
	Participants map[int64]bool
}

// currentEvent — глобальная переменная для активного события.
var currentEvent *AttendanceEvent
