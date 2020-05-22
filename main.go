package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 3400, "The port the server will run on.")
	addr := flag.String("address", "0.0.0.0", "The address the server will run on.")
	flag.Parse()
	dirToServe := flag.Arg(0)
	if len(dirToServe) == 0 {
		dirToServe = "./"
	}

	http.Handle("/", http.FileServer(http.Dir(dirToServe)))

	ip := fmt.Sprintf("%s:%v", *addr, *port)
	log.Printf("Serving %v on %v\n", dirToServe, ip)
	err := http.ListenAndServe(ip, nil)
	if err != nil {
		log.Fatal(err)
	}
}
