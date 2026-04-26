package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/handlers"
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
		s.mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(s.hfs)))
		s.mux.Handle("/post/", s.makeHfsInjector(handlers.PostHandler))
		s.mux.HandleFunc("/", s.makeHfsInjector(handlers.GalleryHandler))
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
	fmt.Printf("Serving %v on %v\n", s.config.Directory, addr)
	return http.ListenAndServe(addr, s.mux)
}
