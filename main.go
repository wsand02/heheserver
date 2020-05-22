package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

const (
	portDesc    string = "The port the server will run on."
	addrDesc    string = "The address the server will run on."
	defaultPort int    = 3400
	defaultDir  string = "./"
	defaultAddr string = "0.0.0.0"
)

func main() {
	// Define long flags
	port := flag.Int("port", defaultPort, portDesc)
	addr := flag.String("address", defaultAddr, addrDesc)
	// Define short flags
	flag.IntVar(port, "p", defaultPort, portDesc)
	flag.StringVar(addr, "a", defaultAddr, addrDesc)
	flag.Parse()

	dirToServe := flag.Arg(0)
	if len(dirToServe) == 0 {
		dirToServe = defaultDir
	}

	http.Handle("/", http.FileServer(http.Dir(dirToServe)))

	ip := fmt.Sprintf("%s:%v", *addr, *port)

	log.Printf("Serving %v on %v\n", dirToServe, ip)
	err := http.ListenAndServe(ip, nil)
	if err != nil {
		log.Fatal(err)
	}
}
