package auth

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)


var DB *sql.DB

func InitDB() {
	os.MkdirAll("data", 0755)
	var err error
	DB, err = sql.Open("sqlite3", "data/codestash.db")
	if err != nil {
		fmt.Println("Error while connecting to DB")
		return
	}
	DB.Exec(`CREATE TABLE IF NOT EXISTS users (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	username TEXT UNIQUE NOT NULL,
    	email TEXT UNIQUE NOT NULL,
    	password_hash TEXT NOT NULL,
    	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	
	DB.Exec(`CREATE TABLE IF NOT EXISTS repo_access (
    	repo_name TEXT NOT NULL,
    	username TEXT NOT NULL,
    	role TEXT NOT NULL,
    	PRIMARY KEY (repo_name, username)
		)`)
	
	DB.Exec(`CREATE TABLE IF NOT EXISTS sessions (
    	token TEXT PRIMARY KEY,
    	username TEXT NOT NULL,
    	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    	expires_at DATETIME NOT NULL
		)`)

}