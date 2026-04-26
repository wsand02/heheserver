package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/wsand02/heheserver/internal/utils"
)

const (
	portDesc                 string = "The port the server will run on."
	hostDesc                 string = "The host the server will run on."
	defaultPort              int    = 3400
	defaultDir               string = "./"
	defaultHost              string = "0.0.0.0"
	galleryDesc              string = "Enables the embedded gallery page."
	defaultGalleryFlag       bool   = false
	resizeDesc               string = "Enables the experimental image resizing endpoint, requires ffmpeg on path."
	defaultSplit             int    = 64
	splitDesc                string = "Max items per page for gallery pagination."
	defaultResizeFlag        bool   = false
	defaultIgnoreCacheSize   int64  = 16
	ignoreCacheSizeDesc      string = "Size of ignore cache in megabytes."
	defaultResizeCacheSize   int64  = 1000
	resizeCacheSizeDesc      string = "Size of resize cache in megabytes."
	defaultVidThumbCacheSize int64  = 1000
	vidThumbCacheSizeDesc    string = "Size of video thumbnail cache in megabytes."
)

type Config struct {
	Port              int
	Host              string
	Gallery           bool
	Resize            bool
	Directory         string
	Split             int
	FFmpegExists      bool
	IgnoreCacheSize   int64
	VidThumbCacheSize int64
	ResizeCacheSize   int64
}

func NewConfig(port, split int, gallery, resize bool, directory, host string, igncsize, rszcsize, vidthmbcsize int64) (*Config, error) {
	if split < 1 {
		return nil, fmt.Errorf("Split has to be greater than 0")
	}
	// prevents infinite loop incase i accidentally add a slash to directory on windows
	_, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("directory does not exist: %w", err)
	}

	return &Config{
		Port:              port,
		Host:              host,
		Directory:         directory,
		Gallery:           gallery,
		Resize:            resize,
		Split:             split,
		FFmpegExists:      utils.FFmpegExists(),
		IgnoreCacheSize:   igncsize,
		VidThumbCacheSize: vidthmbcsize,
		ResizeCacheSize:   rszcsize,
	}, nil
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%v", c.Host, c.Port)
}

func ParseFromFlags() (*Config, error) {
	port := flag.Int("port", defaultPort, portDesc)
	host := flag.String("host", defaultHost, hostDesc)
	gallery := flag.Bool("gallery", defaultGalleryFlag, galleryDesc)
	resize := flag.Bool("resize", defaultResizeFlag, resizeDesc)
	split := flag.Int("split", defaultSplit, splitDesc)
	ignoreCacheSize := flag.Int64("igncache", defaultIgnoreCacheSize, ignoreCacheSizeDesc)
	resizeCacheSize := flag.Int64("rescache", defaultResizeCacheSize, resizeCacheSizeDesc)
	vidThumbCacheSize := flag.Int64("vidtcache", defaultVidThumbCacheSize, vidThumbCacheSizeDesc)
	// Define short flags
	flag.IntVar(port, "p", defaultPort, portDesc)
	flag.StringVar(host, "h", defaultHost, hostDesc)
	flag.IntVar(split, "s", defaultSplit, splitDesc)
	flag.BoolVar(gallery, "g", defaultGalleryFlag, galleryDesc)
	flag.BoolVar(resize, "r", defaultResizeFlag, resizeDesc)
	flag.Parse()

	dirToServe := flag.Arg(0)
	if len(dirToServe) == 0 {
		dirToServe = defaultDir
	}
	return NewConfig(*port, *split, *gallery, *resize, dirToServe, *host, *ignoreCacheSize, *resizeCacheSize, *vidThumbCacheSize)
}
