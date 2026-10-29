package handlers

import (
	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/models"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
)

func parseGalleryFilter(r *http.Request) models.GalleryFilter {
	q := r.URL.Query()
	f := models.GalleryFilter{}

	valid := map[string]bool{"image": true, "video": true, "audio": true, "dir": true, "other": true}
	for _, t := range q["type"] {
		if valid[t] {
			if f.Types == nil {
				f.Types = map[string]bool{}
			}
			f.Types[t] = true
		}
	}

	f.Query = strings.ToLower(strings.TrimSpace(q.Get("q")))

	for _, e := range strings.Split(q.Get("ext"), ",") {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		if f.Exts == nil {
			f.Exts = map[string]bool{}
		}
		f.Exts[e] = true
	}
	return f
}

func probeImageDimensions(hfs *fs.HeheFS, items []models.GalleryItem) {
	dc := cache.GetDimensionCache()
	for i := range items {
		it := &items[i]
		if it.IsDir || !it.IsImage() {
			continue
		}
		full := it.Path + it.Filename
		if pt, ok := dc.Get(full); ok {
			it.Width, it.Height = pt.X, pt.Y
			continue
		}
		f, err := hfs.Open(full)
		if err != nil {
			continue
		}
		cfg, _, err := image.DecodeConfig(f)
		f.Close()
		if err != nil || cfg.Width == 0 || cfg.Height == 0 {
			continue
		}
		it.Width, it.Height = cfg.Width, cfg.Height
		dc.Set(full, image.Pt(cfg.Width, cfg.Height), 1)
	}
}
