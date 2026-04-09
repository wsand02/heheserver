package server

import (
	"fmt"
	"net/http"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/config"
	"github.com/wsand02/heheserver/internal/server/handlers"
)

type Server struct {
	config *config.Config
	mux    *http.ServeMux
}

func newGallery(cfg *config.Config, hfs http.FileSystem) *Server {
	mux := http.NewServeMux()
	mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(hfs)))
	mux.HandleFunc("/", handlers.GalleryHandler)
	return &Server{
		config: cfg,
		mux:    mux,
	}
}

func newFileServer(cfg *config.Config, hfs http.FileSystem) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(hfs))
	return &Server{
		config: cfg,
		mux:    mux,
	}
}

func NewServer(cfg *config.Config) *Server {
	hfs := fs.Dir(cfg.Directory)
	if cfg.Gallery {
		fmt.Println("Gallery Enabled")
		return newGallery(cfg, hfs)
	}
	return newFileServer(cfg, hfs)
}

func (s *Server) Start() error {
	addr := s.config.GetAddress()
	fmt.Printf("Serving %v on %v\n", s.config.Directory, addr)
	return http.ListenAndServe(addr, s.mux)
}
