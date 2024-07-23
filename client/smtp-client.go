package client

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var user1 string = "ojojohnson@localhost.com"
var password1 string = "jeremiah"

var user2 string = "marvelous@localhost.com"
var password2 string = "marvelous"

var recipient = []string{user1}

func createUser(username string, password string) error {
	var db *sql.DB
	var err error

	db, err = sql.Open("sqlite3", "./users.db")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	hashedPassword, err := hashPassword(password)

	if err != nil {
		log.Panic("Unable to get hashed password")
	}

	var query = fmt.Sprintf(`
		INSERT INTO users (email, password) VALUES ('%s', '%s');
	`, username, hashedPassword)

	_, err = db.Exec(query)

	return err
}

func RunClient(wg *sync.WaitGroup, serverReady *chan struct{}) {
	defer wg.Done()
	log.Println("running client, waiting for server to start")

	<-*serverReady

	log.Println("Server started already, creating user")

	if err := createUser(user2, password2); err != nil {
		log.Println("Created User successfully")
	}

	smtpHost := "localhost"
	smtpPort := "1025"

	message := []byte("Subject: Hello brooooo\n\nWhat is this email about? \n Please respond!.")

	auth := smtp.PlainAuth("", user2, password2, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, user2, recipient, message)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Email Sent Successfully!")
}

// Helper function to hash passwords
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
