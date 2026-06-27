package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(username, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = DB.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)", username, email, string(hash))
	if err != nil {
		return err
	}
	return nil
}

func AuthenticateUser(username, password string) bool {
	var hash string
	err := DB.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true
}

func GetUser(username string) (string, string, error) {
	var uname, email string
	err := DB.QueryRow("SELECT username, email FROM users WHERE username = ?", username).Scan(&uname, &email)
	return uname, email, err
}

