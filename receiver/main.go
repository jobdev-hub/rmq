package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/receiver/test", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		log.Printf("MSG: %s", body)
	})
	log.Fatal(http.ListenAndServe(":8081", nil))
}
