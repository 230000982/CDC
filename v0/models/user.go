package models

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user record
type User struct {
	ID       int
	Nome     string
	Email    string
	Password string
	CargoID  int
}

// Cargo represents a cargo (role) record
type Cargo struct {
	ID        int
	Descricao string
}

// Authenticate checks if the provided credentials are valid
func Authenticate(db *sql.DB, email, password string) (int, int, error) {
	var id, cargo int
	var storedPassword string

	err := db.QueryRow("SELECT id_user, password, cargo_id FROM user WHERE email = ?", email).Scan(&id, &storedPassword, &cargo)
	if err != nil {
		return 0, 0, err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return 0, 0, err
	}

	return id, cargo, nil
}

// CreateUser creates a new user
func CreateUser(db *sql.DB, nome, email, password string, cargoID int) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insert the new user
	_, err = db.Exec("INSERT INTO user (nome, email, password, cargo_id) VALUES (?, ?, ?, ?)",
		nome, email, string(hashedPassword), cargoID)

	return err
}

// EmailExists checks if an email already exists in the database
func EmailExists(db *sql.DB, email string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, id int) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id_user, nome, email, cargo_id FROM user WHERE id_user = ?", id).Scan(
		&user.ID, &user.Nome, &user.Email, &user.CargoID,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
func UpdateUser(db *sql.DB, user *User) error {
	_, err := db.Exec("UPDATE user SET nome = ?, email = ?, cargo_id = ? WHERE id_user = ?",
		user.Nome, user.Email, user.CargoID, user.ID)

	return err
}

// UpdatePassword updates a user's password
func UpdatePassword(db *sql.DB, userID int, newPassword string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update the password
	_, err = db.Exec("UPDATE user SET password = ? WHERE id_user = ?", string(hashedPassword), userID)

	return err
}

// DeleteUser deletes a user by ID
func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM user WHERE id_user = ?", id)
	return err
}

// GetAllUsers retrieves all users
func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id_user, nome, email, cargo_id FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Nome, &user.Email, &user.CargoID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetAllCargos retrieves all cargos (roles)
func GetAllCargos(db *sql.DB) ([]Cargo, error) {
	rows, err := db.Query("SELECT id_cargo, descricao FROM cargo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cargos []Cargo
	for rows.Next() {
		var cargo Cargo
		if err := rows.Scan(&cargo.ID, &cargo.Descricao); err != nil {
			return nil, err
		}
		cargos = append(cargos, cargo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cargos, nil
}
