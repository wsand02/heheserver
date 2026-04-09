package server

import (
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/config"
	"github.com/wsand02/heheserver/internal/server/handlers"
)

type Server struct {
	config *config.Config
	mux    *http.ServeMux
}

func NewServer(cfg *config.Config) (*Server, error) {
	mux := http.NewServeMux()
	hfs := fs.Dir(cfg.Directory)
	if cfg.Gallery {
		mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(hfs)))
		mux.HandleFunc("/", handlers.GalleryHandler)
	} else {
		mux.Handle("/", http.FileServer(hfs))
	}
	return &Server{
		config: &config.Config{},
		mux:    mux,
	}, nil
}

func (s *Server) Start() error {
	addr := s.config.GetAddress()
	log.Printf("Serving %v on %v\n", s.config.Directory, addr)
	return http.ListenAndServe(addr, s.mux)
}
