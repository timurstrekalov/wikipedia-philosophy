package main

import (
	"encoding/json"
	"github.com/timurstrekalov/wikipedia-philosophy/wikipedia"
	"log"
	"net/http"
)

const startUrl = "Special:Random"

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/path", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		path, err := wikipedia.GetPath(from, to)
		if err != nil {
			log.Printf("error getting path from %s to %s: %s", from, to, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(path)
		if err != nil {
			log.Printf("error encoding JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
