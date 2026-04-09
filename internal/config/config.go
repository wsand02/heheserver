package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	portDesc           string = "The port the server will run on."
	addrDesc           string = "The address the server will run on."
	defaultPort        int    = 3400
	defaultDir         string = "./"
	defaultAddr        string = "0.0.0.0"
	galleryDesc        string = "Enables the embedded gallery page. Which currently uses ThinGallery."
	defaultGalleryFlag bool   = false
)

type Config struct {
	Port      int
	Address   string
	Gallery   bool
	Directory string
}

func ParseFromFlags() (*Config, error) {
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
		return nil, fmt.Errorf("directory does not exist: %w", err)
	}

	return &Config{
		Port:      *port,
		Address:   *addr,
		Directory: dirToServe,
		Gallery:   *gallery,
	}, nil
}
