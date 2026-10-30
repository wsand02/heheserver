package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/frontend"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/handlers"
	"github.com/wsand02/heheserver/internal/thumbgen"
)

type Server struct {
	config *config.Config
	mux    *http.ServeMux
	hfs    *fs.HeheFS
}

func (s *Server) makeHfsInjector(fn func(http.ResponseWriter, *http.Request, string, *fs.HeheFS, *config.Config)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("path")
		if q == "" {
			q = "/"
		}
		fn(w, r, q, s.hfs, s.config)
	}
}

func (s *Server) initCache() {
	err := cache.NewIgnoreCache(s.config.IgnoreCacheSize)
	if err != nil {
		log.Fatal(err)
	}
	if s.config.Gallery {
		if err = cache.NewDimensionCache(); err != nil {
			log.Fatal(err)
		}
	}
	if s.config.Resize {
		err = cache.NewResizeCache(s.config.ResizeCacheSize)
		if err != nil {
			log.Fatal(err)
		}
		if s.config.FFmpegExists {
			err = cache.NewVidThumbCache(s.config.VidThumbCacheSize)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (s *Server) setupRoutes() {
	if s.config.Gallery {
		fmt.Println("Enabling Gallery")
		s.mux.HandleFunc("/api/", handlers.UnknownAPI)
		s.mux.Handle("GET /api/gallery", s.makeHfsInjector(handlers.GalleryAPI))
		s.mux.Handle("GET /api/media", s.makeHfsInjector(handlers.MediaAPI))
		s.mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(s.hfs)))
		s.mux.Handle("/", frontend.Handler())
		if s.config.Resize {
			fmt.Println("Enabling Resize Endpoint")
			s.mux.HandleFunc("/resize/", s.makeHfsInjector(handlers.ResizeHandler))
			if s.config.FFmpegExists {
				fmt.Println("FFmpeg found, Enabling Video Thumbnail Endpoint")
				s.mux.HandleFunc("/vidthumb/", s.makeHfsInjector(handlers.VidThumbHandler))
			} else {
				fmt.Println("FFmpeg not found")
				fmt.Println("Will resize using fallback")
			}

		}

		return
	}
	s.mux.Handle("/", http.FileServer(s.hfs))
}

func NewServer(cfg *config.Config) *Server {
	workers := cfg.ThumbWorkers
	// Keep zero-value Config literals compatible with the original default.
	if workers == 0 {
		workers = thumbgen.DefaultMaxConcurrent
	}
	thumbgen.Configure(workers)
	mux := http.NewServeMux()
	hfs := fs.Dir(cfg.Directory)
	hfsa, ok := hfs.(*fs.HeheFS)
	if !ok {
		log.Fatal("not hehefs")
	}
	srv := &Server{
		config: cfg,
		mux:    mux,
		hfs:    hfsa,
	}
	srv.initCache()
	srv.setupRoutes()
	return srv
}

func (s *Server) Start() error {
	addr := s.config.GetAddress()
	fmt.Printf("Serving %v\n", s.config.Directory)
	for _, url := range s.config.GetDisplayURLs() {
		fmt.Printf("  %s\n", url)
	}
	return http.ListenAndServe(addr, s.mux)
}
