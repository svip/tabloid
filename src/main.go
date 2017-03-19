package main

import (
	"net/http"
	"log"
	"pages"
)

func main() {
	http.HandleFunc("/", pages.HeadlinePage)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("media/"))))
	
	log.Fatal(http.ListenAndServe(":8070", nil))
}
