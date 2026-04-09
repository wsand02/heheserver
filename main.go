package main

import (
	"log"

	"github.com/wsand02/heheserver/internal/server"
	"github.com/wsand02/heheserver/internal/server/config"
	"github.com/wsand02/heheserver/internal/version"
)

func main() {
	version.PrintVersion()
	config, err := config.ParseFromFlags()
	if err != nil {
		log.Fatal(err)
	}
	s, err := server.NewServer(config)
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
