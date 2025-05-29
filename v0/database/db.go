package database

import (
	"database/sql"
	"log"
	"v0/config"

	_ "github.com/go-sql-driver/mysql"
)

// Connect establishes a connection to the database
func Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Database connection established")
	return db, nil
}

// GetTipos retrieves all tipos from the database
func GetTipos(db *sql.DB) ([]struct {
	ID        int
	Descricao string
}, error) {
	rows, err := db.Query("SELECT id_tipo, descricao FROM tipo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tipos []struct {
		ID        int
		Descricao string
	}
	for rows.Next() {
		var t struct {
			ID        int
			Descricao string
		}
		if err := rows.Scan(&t.ID, &t.Descricao); err != nil {
			return nil, err
		}
		tipos = append(tipos, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tipos, nil
}

// GetPlataformas retrieves all plataformas from the database
func GetPlataformas(db *sql.DB) ([]struct {
	ID        int
	Descricao string
}, error) {
	rows, err := db.Query("SELECT id_platforma, descricao FROM plataforma")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plataformas []struct {
		ID        int
		Descricao string
	}
	for rows.Next() {
		var p struct {
			ID        int
			Descricao string
		}
		if err := rows.Scan(&p.ID, &p.Descricao); err != nil {
			return nil, err
		}
		plataformas = append(plataformas, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return plataformas, nil
}

// GetEstados retrieves all estados from the database
func GetEstados(db *sql.DB) ([]struct {
	ID        int
	Descricao string
}, error) {
	rows, err := db.Query("SELECT id_estado, descricao FROM estado")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var estados []struct {
		ID        int
		Descricao string
	}
	for rows.Next() {
		var e struct {
			ID        int
			Descricao string
		}
		if err := rows.Scan(&e.ID, &e.Descricao); err != nil {
			return nil, err
		}
		estados = append(estados, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return estados, nil
}

// GetResultados retrieves all resultados from the database
func GetResultados(db *sql.DB) ([]struct {
	ID        int
	Descricao string
}, error) {
	rows, err := db.Query("SELECT id_resultado, descricao FROM resultado")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resultados []struct {
		ID        int
		Descricao string
	}
	for rows.Next() {
		var r struct {
			ID        int
			Descricao string
		}
		if err := rows.Scan(&r.ID, &r.Descricao); err != nil {
			return nil, err
		}
		resultados = append(resultados, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return resultados, nil
}

// GetEmailsFromDB retrieves emails from the database for notification
func GetEmailsFromDB(db *sql.DB) ([]string, error) {
	var emails []string

	rows, err := db.Query(`
        SELECT email FROM user 
        WHERE cargo_id NOT IN (1, 4) OR cargo_id IS NULL
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}

// GetAllEmails retrieves all emails from the user table
func GetAllEmails(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT email FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}
