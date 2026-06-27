package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func CreateSession(username string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	expires := time.Now().Add(24 * time.Hour)
	_, err := DB.Exec("INSERT INTO sessions (token, username, expires_at) VALUES (?, ?, ?)", token, username, expires)
	if err != nil {
		return "", err
	}
	return token, nil
}

func CleanExpiredSessions() {
	DB.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
}

func DeleteSession(token string) {
	if token == "" {
        return
    }
    DB.Exec("DELETE FROM sessions WHERE token = ?", token)
}

func GetSession(token string) string {
	if token == "" {
		return ""
	}
	var username string
	err := DB.QueryRow("SELECT username FROM sessions WHERE token = ? AND expires_at > ?", token, time.Now()).Scan(&username)
	if err != nil {
		return ""
	}
	return username
}