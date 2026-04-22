package main

import (
	"fmt"
	"log"

	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/server"
	"github.com/wsand02/heheserver/internal/version"
)

func main() {
	fmt.Printf("heheserver %s\n", version.GetVersion())
	config, err := config.ParseFromFlags()
	if err != nil {
		log.Fatal(err)
	}
	s := server.NewServer(config)
	s.Start()
}
