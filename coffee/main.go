package main

import (
	"log"
	"net/http"
	"os"
)

func coffeeHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Your Coffee has been served by - " + servant))
}

func pourOverHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Your Pour Over has been served by - " + servant))
}

func aeropressHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Your Aeropress has been served by - " + servant))
}

func main() {
	http.HandleFunc("/coffee", coffeeHandler)
	http.HandleFunc("/coffee/pour-over", pourOverHandler)
	http.HandleFunc("/coffee/aeropress", aeropressHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
