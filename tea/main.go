package main

import (
	"log"
	"net/http"
	"os"
)

func teaHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Your Tea has been served by - " + servant))
}

func main() {
	http.HandleFunc("/tea", teaHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
