package auth

import (
	"net/http"
)

func GetCurrentUser(r *http.Request) string {
	session, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	username := GetSession(session.Value)
	return username
}