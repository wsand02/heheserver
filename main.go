package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/wsand02/heheserver/internal"
)

//go:embed templates
var templates embed.FS

const (
	portDesc           string = "The port the server will run on."
	addrDesc           string = "The address the server will run on."
	defaultPort        int    = 3400
	defaultDir         string = "./"
	defaultAddr        string = "0.0.0.0"
	galleryDesc        string = "Enables the embedded gallery page. Which currently uses ThinGallery."
	defaultGalleryFlag bool   = false
)

func main() {
	// Define long flags
	port := flag.Int("port", defaultPort, portDesc)
	addr := flag.String("address", defaultAddr, addrDesc)
	gallery := flag.Bool("gallery", defaultGalleryFlag, galleryDesc)
	// Define short flags
	flag.IntVar(port, "p", defaultPort, portDesc)
	flag.StringVar(addr, "a", defaultAddr, addrDesc)
	flag.BoolVar(gallery, "g", defaultGalleryFlag, galleryDesc)
	flag.Parse()

	dirToServe := flag.Arg(0)
	if len(dirToServe) == 0 {
		dirToServe = defaultDir
	}

	// prevents infinite loop incase i accidentally add a slash to directory on windows
	_, err := os.Stat(dirToServe)
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	if *gallery {
		log.Println("Embedded Gallery enabled")
		http.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(internal.Dir(dirToServe))))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			err := tmpl.ExecuteTemplate(w, "gallery.html", nil)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Printf("Error executing template: %v\n", err)
				return
			}
		})
	} else {
		http.Handle("/", http.FileServer(internal.Dir(dirToServe)))
	}

	ip := fmt.Sprintf("%s:%v", *addr, *port)

	log.Printf("Serving %v on %v\n", dirToServe, ip)
	err = http.ListenAndServe(ip, nil)
	if err != nil {
		log.Fatal(err)
	}
}
