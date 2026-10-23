package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

type ScoreEntry struct {
	PlayerName string `json:"name"`
	Score      int    `json:"score"`
}

var (
	ranking []ScoreEntry
	mutex   sync.Mutex
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func rankingHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	switch r.Method {
	case "GET":
		limit := 10
		if len(ranking) < limit {
			limit = len(ranking)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ranking[:limit])

	case "POST":
		var entry ScoreEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ranking = append(ranking, entry)
		sort.Slice(ranking, func(i, j int) bool {
			return ranking[i].Score > ranking[j].Score
		})

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"status":"success"}`)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/ranking", enableCORS(rankingHandler))

	log.Printf("Server started on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}