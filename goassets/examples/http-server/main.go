package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

// make assets to build assets
// make run to build assets and start server
// GOASSETS_PATH=assets make run to start server and always read local assets

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// make all files in assets accessible via /static/<name>, e.g. /static/style.css
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(FileSystem())))

	// root handler for layout
	http.HandleFunc("/", handler)
	e := http.ListenAndServe(":"+port, nil)
	if e != nil {
		log.Fatal(e)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	f, e := FileSystem().Open("layout.html")
	if e != nil {
		http.NotFound(w, r)
		return
	}
	_, e = io.Copy(w, f)
	if e != nil {
		log.Printf("ERROR: %q", e)
	}
	return
}
