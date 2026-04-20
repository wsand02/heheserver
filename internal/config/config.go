package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	portDesc           string = "The port the server will run on."
	hostDesc           string = "The host the server will run on."
	defaultPort        int    = 3400
	defaultDir         string = "./"
	defaultHost        string = "0.0.0.0"
	galleryDesc        string = "Enables the embedded gallery page. Which currently uses ThinGallery."
	defaultGalleryFlag bool   = false
	resizeDesc         string = "Enables the experimental image resizing endpoint"
	defaultResizeFlag  bool   = false
)

type Config struct {
	Port      int
	Host      string
	Gallery   bool
	Resize    bool
	Directory string
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%v", c.Host, c.Port)
}

func ParseFromFlags() (*Config, error) {
	port := flag.Int("port", defaultPort, portDesc)
	host := flag.String("host", defaultHost, hostDesc)
	gallery := flag.Bool("gallery", defaultGalleryFlag, galleryDesc)
	resize := flag.Bool("resize", defaultResizeFlag, resizeDesc)
	// Define short flags
	flag.IntVar(port, "p", defaultPort, portDesc)
	flag.StringVar(host, "h", defaultHost, hostDesc)
	flag.BoolVar(gallery, "g", defaultGalleryFlag, galleryDesc)
	flag.BoolVar(resize, "r", defaultResizeFlag, resizeDesc)
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
		Host:      *host,
		Directory: dirToServe,
		Gallery:   *gallery,
		Resize:    *resize,
	}, nil
}
