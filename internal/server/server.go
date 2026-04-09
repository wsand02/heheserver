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

func newGallery(cfg *config.Config, mux *http.ServeMux, hfs *fs.HeheFS) *Server {
	mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(hfs)))
	srv := &Server{
		mux:    mux,
		config: cfg,
		hfs:    hfs,
	}
	srv.mux.HandleFunc("/", srv.makeHfsInjector(handlers.GalleryHandler))
	return srv
}

func newFileServer(cfg *config.Config, mux *http.ServeMux, hfs *fs.HeheFS) *Server {
	mux.Handle("/", http.FileServer(hfs))
	return &Server{
		mux:    mux,
		config: cfg,
		hfs:    hfs,
	}
}

func NewServer(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	hfs := fs.Dir(cfg.Directory)
	hfsa, ok := hfs.(*fs.HeheFS)
	if !ok {
		log.Fatal("not hehefs")
	}
	if cfg.Gallery {
		fmt.Println("Gallery Enabled")
		return newGallery(cfg, mux, hfsa)
	}
	return newFileServer(cfg, mux, hfsa)
}

func (s *Server) Start() error {
	addr := s.config.GetAddress()
	fmt.Printf("Serving %v on %v\n", s.config.Directory, addr)
	return http.ListenAndServe(addr, s.mux)
}
