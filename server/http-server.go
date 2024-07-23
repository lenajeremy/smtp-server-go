package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Controll-Allow-Methods", "GET, POST, OPTIONS")
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type Mail = map[string]string
type MailObject struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

func SetupHTTPServer(wg *sync.WaitGroup) {
	defer wg.Done()

	r := mux.NewRouter()

	r.Use(enableCors)
	r.Use(requestLogger)

	r.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	r.HandleFunc("/mails", func(w http.ResponseWriter, r *http.Request) {
		mails := []MailObject{}

		rows, err := DB.Query(`
			SELECT "from_email", "to_email", subject, body, created_at 
			FROM emails 
			ORDER BY created_at DESC 
			LIMIT 20
		`)

		if err != nil {
			log.Fatal(err)
		}

		defer rows.Close()

		// Process the rows
		for rows.Next() {
			var m MailObject
			err := rows.Scan(&m.From, &m.To, &m.Subject, &m.Body, &m.CreatedAt)

			if err != nil {
				log.Fatal(err)
			}

			mails = append(mails, m)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, "Error after reading rows", http.StatusInternalServerError)
			log.Printf("Rows error: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(mails); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			log.Printf("JSON encoding error: %v", err)
		}
	}).Methods("GET")

	r.HandleFunc("/mail/send", func(w http.ResponseWriter, r *http.Request) {
		type RequestBody struct {
			From    string `json:"from"`
			To      string `json:"to"`
			Body    string `json:"body"`
			Subject string `json:"subject"`
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var body RequestBody

		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		InsertEmail(body.From, body.To, body.Subject, body.Body)

		json.NewEncoder(w).Encode(body)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dict := map[string]any{"name": "jeremiah", "age": 22, "sings": true, "dances": false}
		json.NewEncoder(w).Encode(dict)
	})

	log.Println("HTTP server running on localhost:8000")
	http.ListenAndServe(":8000", r)
}
