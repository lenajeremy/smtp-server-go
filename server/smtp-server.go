package server

import (
	"database/sql"
	"io"
	"log"
	"sync"
	"time"

	"jeremiah.smtp/utils"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func initDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}

	// SQL statements to create tables
	sqlStatements := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS emails (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        from_email TEXT NOT NULL,
        to_email TEXT NOT NULL,
        subject TEXT,
        body TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `

	_, err = DB.Exec(sqlStatements)

	if err != nil {
		log.Fatal(err)
	}
}

// The Backend implements SMTP server methods.
type Backend struct{}

// NewSession is called after client greeting (EHLO, HELO).
func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}

// Session represents a user session.
type Session struct {
	From string
	To   string
}

// AuthMechanisms returns a slice of available auth mechanisms; only PLAIN is
// supported in this example.
func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

// Auth is the handler for supported authenticators.
func (s *Session) Auth(mech string) (sasl.Server, error) {
	return sasl.NewPlainServer(func(identity, username, password string) error {
		var hashedPassword string
		err := DB.QueryRow("SELECT password FROM users WHERE email = ?", username).Scan(&hashedPassword)

		if err != nil {
			log.Printf(err.Error())
			return smtp.ErrAuthRequired
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

		if err != nil {
			return smtp.ErrAuthFailed
		}

		// the username and password matches and there's no error
		return nil
	}), nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.From = from
	log.Println("Mail from:", from)
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	if s.From != to {
		s.To = to
		log.Println("Mail to:", to)
	}
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := io.ReadAll(r); err != nil {
		return err
	} else {
		msg := string(b)
		subject, body := utils.ParseEmail(msg)

		err = InsertEmail(s.From, s.To, subject, body)
		if err != nil {
			log.Println(err)
		}
		log.Println("Data:", string(b))
	}
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

func InsertEmail(from, to, subject, body string) error {
	_, err := DB.Exec(`
        INSERT INTO emails (from_email, to_email, subject, body)
        VALUES (?, ?, ?, ?)
    `, from, to, subject, body)
	return err
}

func SetupSMTPServer(wg *sync.WaitGroup, ch *chan struct{}) {
	defer func() {
		DB.Close()
		wg.Done()
	}()

	log.Println("Initializing Server")

	initDB()

	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = "localhost:1025"
	s.Domain = "localhost"
	s.WriteTimeout = 10 * time.Second
	s.ReadTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("Starting server at", s.Addr)
	close(*ch)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}