package main

import (
	"log"
	"net/http"
	"os"
)

func teaHandler(w http.ResponseWriter, r *http.Request) {
	servant, err := os.Hostname()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error, no Tea for your :("))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Your Tea has been served by - " + servant))
}

func main() {
	http.HandleFunc("/tea", teaHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
