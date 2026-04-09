package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/fs"
)

func main() {

	if *gallery {
		log.Println("Embedded Gallery enabled")
		http.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(fs.Dir(dirToServe))))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			err := tmpl.ExecuteTemplate(w, "gallery.html", nil)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Printf("Error executing template: %v\n", err)
				return
			}
		})
	} else {
		http.Handle("/", http.FileServer(fs.Dir(dirToServe)))
	}

	ip := fmt.Sprintf("%s:%v", *addr, *port)

	log.Printf("Serving %v on %v\n", dirToServe, ip)
	err = http.ListenAndServe(ip, nil)
	if err != nil {
		log.Fatal(err)
	}
}
