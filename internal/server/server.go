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
	hfs    fs.HeheFS
}

func (s *Server) GetHfs() fs.HeheFS {
	return s.hfs
}

func (s *Server) makeHfsHandler(fn func(http.ResponseWriter, *http.Request, any)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := handlers.BuildContext(s.hfs)
		fn(w, r, ctx)
	}
}

func newGallery(cfg *config.Config, hfs http.FileSystem, hehefs fs.HeheFS) *Server {
	mux := http.NewServeMux()
	mux.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(hfs)))
	srv := &Server{
		config: cfg,
		mux:    mux,
		hfs:    hehefs,
	}
	mux.HandleFunc("/", srv.makeHfsHandler(handlers.GalleryHandler))
	return srv
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
	heheFs, ok := hfs.(*fs.HeheFS) // *HeheFS
	if !ok {
		log.Fatal("hfs not hehefs")
	}
	if cfg.Gallery {
		fmt.Println("Gallery Enabled")
		return newGallery(cfg, hfs, *heheFs)
	}
	return newFileServer(cfg, hfs)
}

func (s *Server) Start() error {
	addr := s.config.GetAddress()
	fmt.Printf("Serving %v on %v\n", s.config.Directory, addr)
	return http.ListenAndServe(addr, s.mux)
}
