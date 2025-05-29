package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

// LogService handles logging operations
type LogService struct {
	db *sql.DB
}

// NewLogService creates a new LogService
func NewLogService(db *sql.DB) *LogService {
	return &LogService{
		db: db,
	}
}

// LogAction logs an action to the database
func (s *LogService) LogAction(userID int, table string, action string, oldData interface{}, newData interface{}) error {
	// Convert old data to JSON if not nil
	var oldDataJSON sql.NullString
	if oldData != nil {
		oldJSON, err := json.Marshal(oldData)
		if err != nil {
			log.Printf("Error marshaling old data: %v", err)
			return err
		}
		oldDataJSON = sql.NullString{String: string(oldJSON), Valid: true}
	}

	// Convert new data to JSON if not nil
	var newDataJSON sql.NullString
	if newData != nil {
		newJSON, err := json.Marshal(newData)
		if err != nil {
			log.Printf("Error marshaling new data: %v", err)
			return err
		}
		newDataJSON = sql.NullString{String: string(newJSON), Valid: true}
	}

	// Insert log entry
	_, err := s.db.Exec(
		"INSERT INTO logs (tabela, acao, old_data, new_data, id_user) VALUES (?, ?, ?, ?, ?)",
		table,
		action,
		oldDataJSON,
		newDataJSON,
		userID,
	)

	if err != nil {
		log.Printf("Error inserting log: %v", err)
		return err
	}

	return nil
}

// LogCreate logs a create action
func (s *LogService) LogCreate(userID int, table string, data interface{}) error {
	return s.LogAction(userID, table, "create", nil, data)
}

// LogUpdate logs an update action
func (s *LogService) LogUpdate(userID int, table string, oldData interface{}, newData interface{}) error {
	return s.LogAction(userID, table, "update", oldData, newData)
}

// LogDelete logs a delete action
func (s *LogService) LogDelete(userID int, table string, data interface{}) error {
	return s.LogAction(userID, table, "delete", data, nil)
}

// GetLogs retrieves logs from the database with optional filtering
func (s *LogService) GetLogs(table string, action string, userID int, limit int, offset int) ([]map[string]interface{}, error) {
	// Build query with optional filters
	query := "SELECT id_logs, tabela, acao, old_data, new_data, timestamp, id_user FROM logs WHERE 1=1"
	var args []interface{}

	if table != "" {
		query += " AND tabela = ?"
		args = append(args, table)
	}

	if action != "" {
		query += " AND acao = ?"
		args = append(args, action)
	}

	if userID > 0 {
		query += " AND id_user = ?"
		args = append(args, userID)
	}

	// Add ordering and pagination
	query += " ORDER BY timestamp DESC"
	
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying logs: %w", err)
	}
	defer rows.Close()

	// Process results
	var logs []map[string]interface{}
	for rows.Next() {
		var id int
		var table, action string
		var oldData, newData sql.NullString
		var timestamp string
		var userID int

		if err := rows.Scan(&id, &table, &action, &oldData, &newData, &timestamp, &userID); err != nil {
			return nil, fmt.Errorf("error scanning log row: %w", err)
		}

		// Create log entry
		logEntry := map[string]interface{}{
			"id":        id,
			"table":     table,
			"action":    action,
			"timestamp": timestamp,
			"user_id":   userID,
		}

		// Add old data if valid
		if oldData.Valid {
			var oldDataObj interface{}
			if err := json.Unmarshal([]byte(oldData.String), &oldDataObj); err == nil {
				logEntry["old_data"] = oldDataObj
			} else {
				logEntry["old_data"] = oldData.String
			}
		}

		// Add new data if valid
		if newData.Valid {
			var newDataObj interface{}
			if err := json.Unmarshal([]byte(newData.String), &newDataObj); err == nil {
				logEntry["new_data"] = newDataObj
			} else {
				logEntry["new_data"] = newData.String
			}
		}

		logs = append(logs, logEntry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating log rows: %w", err)
	}

	return logs, nil
}
