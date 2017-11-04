package main

import (
	"log"
	"net/http"
	"pages"
)

func main() {
	// API
	http.HandleFunc("/api", pages.Api)

	// Main entry.
	http.HandleFunc("/", pages.HeadlinePage)

	// You should run it from the base directory, like say:
	// $ go run src/main.go
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("media/"))))

	// Hardcoded some port.
	log.Fatal(http.ListenAndServe(":8070", nil))
}
