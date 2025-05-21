package models

import (
	"database/sql"
)

// NullString is a wrapper around sql.NullString
type NullString struct {
	sql.NullString
}

// ParseNullString converts a string to a NullString
func ParseNullString(s string) NullString {
	if s == "" {
		return NullString{sql.NullString{Valid: false}}
	}
	return NullString{sql.NullString{String: s, Valid: true}}
}
