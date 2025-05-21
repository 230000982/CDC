package utils

import (
	"database/sql"
	"time"
)

// ParseNullString converts a string to a sql.NullString
func ParseNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// IsFutureDate checks if a date and time are in the future
func IsFutureDate(dateStr, timeStr string) bool {
	// Combine date and time strings
	fullStr := dateStr + " " + timeStr

	// Parse layout must match the format of fullStr
	layout := "2006-01-02 15:04:05"
	inputTime, err := time.Parse(layout, fullStr)
	if err != nil {
		// If parsing fails, assume it's not in the future
		return false
	}

	// Compare with current time
	return inputTime.After(time.Now())
}

// CalculateDaysRemaining returns the number of full days from today until the target date
func CalculateDaysRemaining(date string) int {
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return -1
	}

	// Zerar hora de hoje e do target para comparar apenas datas
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	diff := targetDate.Sub(today)
	return int(diff.Hours() / 24)
}

// GetObjetoString returns the string representation of an objeto ID
func GetObjetoString(objetoID int) string {
	objetoMap := map[int]string{
		1: "",
		2: "CTE",
		3: "CON",
		4: "INF",
		5: "CI",
		6: "ROB",
	}

	if str, ok := objetoMap[objetoID]; ok {
		return str
	}
	return ""
}

// GetEstadoString returns the string representation of an estado ID
func GetEstadoString(estadoID int) string {
	estadoMap := map[int]string{
		1: "",
		2: "Em Andamento",
		3: "Enviado",
		4: "Não Enviado",
		5: "Declaração",
	}

	if str, ok := estadoMap[estadoID]; ok {
		return str
	}
	return "N/A"
}
