package web

import (
	"net/http"

	"github.com/NiHaLOO7/codestash/auth"
)

func handleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{
			"Title":      "Sign Up",
			"Error":      "",
			"HideChrome": true,
		}
		renderTemplate(w, "signup.html", data)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if len(password) < 8 {
		data := map[string]interface{}{
			"Title":      "Sign Up",
			"Error":      "Password must be at least 8 characters.",
			"HideChrome": true,
		}
		renderTemplate(w, "signup.html", data)
		return
	}

	if password != confirmPassword {
		data := map[string]interface{}{
			"Title":      "Sign Up",
			"Error":      "Passwords do not match.",
			"HideChrome": true,
		}
		renderTemplate(w, "signup.html", data)
		return
	}

	err := auth.CreateUser(username, email, password)
	if err != nil {
		data := map[string]interface{}{
			"Title":      "Sign Up",
			"Error":      "Username or email already taken.",
			"HideChrome": true,
		}
		renderTemplate(w, "signup.html", data)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{
			"Title":      "Sign In",
			"Error":      "",
			"HideChrome": true,
		}
		renderTemplate(w, "login.html", data)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if !auth.AuthenticateUser(username, password) {
		data := map[string]interface{}{
			"Title":      "Sign In",
			"Error":      "Invalid username or password.",
			"HideChrome": true,
		}
		renderTemplate(w, "login.html", data)
		return
	}

	token, err := auth.CreateSession(username)
	if err != nil {
		data := map[string]interface{}{
			"Title":      "Sign In",
			"Error":      "Something went wrong. Try again.",
			"HideChrome": true,
		}
		renderTemplate(w, "login.html", data)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		auth.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
