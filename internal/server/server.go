package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/config"
	"github.com/wsand02/heheserver/internal/server/handlers"
)

type Server struct {
	config *config.Config
	mux    *http.ServeMux
	hfs    *fs.HeheFS
}

func (s *Server) makeHfsInjector(fn func(http.ResponseWriter, *http.Request, string, *fs.HeheFS)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("path")
		if q == "" {
			q = "/"
		}
		fn(w, r, q, s.hfs)
	}
}

func (s *Server) setupRoutes() {
	if s.config.Gallery {
		s.mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(s.hfs)))
		s.mux.HandleFunc("/", s.makeHfsInjector(handlers.GalleryHandler))
		s.mux.HandleFunc("/resize/", s.makeHfsInjector(handlers.ResizeHandler))
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
	srv.setupRoutes()
	return srv
}

func (s *Server) Start() error {
	addr := s.config.GetAddress()
	fmt.Printf("Serving %v on %v\n", s.config.Directory, addr)
	return http.ListenAndServe(addr, s.mux)
}
