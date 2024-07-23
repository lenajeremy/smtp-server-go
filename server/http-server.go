package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("\n\nhelloooooooo\n\n")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Controll-Allow-Methods", "GET, POST, OPTIONS")
		next.ServeHTTP(w, r)
	})
}

type Mail = map[string]string
type MailObject struct {
	from    string
	to      string
	subject string
	body    string
}

func SetupHTTPServer(wg *sync.WaitGroup) {
	defer wg.Done()

	r := mux.NewRouter()

	r.Use(enableCors)

	r.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	r.HandleFunc("/mails", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Mail{
			{"from": "jeremiah@localhost", "to": "marvelous@localhost", "subject": "how are you", "body": "something interesting"},
			{"from": "marvelous@localhost", "to": "jereamiah@localhost", "subject": "Checking up on you", "body": "I'm doing awesome. Thank you \n\n XOXO😙"},
		})
	})

	r.HandleFunc("/mail/send", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Just hit the send endpoint", r.Method, r.Body)
		var body MailObject

		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("%v", body)

		json.NewEncoder(w).Encode(body)
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dict := map[string]any{"name": "jeremiah", "age": 22, "sings": true, "dances": false}
		json.NewEncoder(w).Encode(dict)
	})

	log.Println("HTTP server running on localhost:8000")
	http.ListenAndServe(":8000", r)
}
