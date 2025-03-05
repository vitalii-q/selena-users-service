package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, user-service!")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // значение по умолчанию, если PORT не указан
	}

	fmt.Printf("Starting server on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
