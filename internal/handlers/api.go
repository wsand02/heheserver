package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/models"
	"github.com/wsand02/heheserver/internal/utils"
	"github.com/wsand02/heheserver/internal/version"
)

var errNotDirectory = errors.New("expected a directory")

type mediaItem struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	URL       string    `json:"url"`
	Thumbnail string    `json:"thumbnail,omitempty"`
	ViewerURL string    `json:"viewerUrl"`
}
type galleryResponse struct {
	Items    []mediaItem `json:"items"`
	Path     string      `json:"path"`
	Page     int         `json:"page"`
	Pages    int         `json:"pages"`
	Total    int         `json:"total"`
	Infinite bool        `json:"infinite"`
	Version  string      `json:"version"`
}
type mediaResponse struct {
	Item       mediaItem `json:"item"`
	Previous   string    `json:"previous,omitempty"`
	Next       string    `json:"next,omitempty"`
	Position   int       `json:"position"`
	Total      int       `json:"total"`
	GalleryURL string    `json:"galleryUrl"`
}

func jsonReply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
func apiError(w http.ResponseWriter, status int, message string) {
	jsonReply(w, status, map[string]any{"error": map[string]any{"message": message, "status": status}})
}

func UnknownAPI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/gallery" || r.URL.Path == "/api/media" {
		w.Header().Set("Allow", "GET, HEAD")
		apiError(w, 405, "Method not allowed")
		return
	}
	apiError(w, 404, "API endpoint not found")
}

// Normalize before consulting ignore rules, and reject Windows-specific escapes.
func apiPath(raw string) (string, error) {
	if strings.ContainsAny(raw, "\\\x00") {
		return "", fmt.Errorf("invalid path")
	}
	for _, part := range strings.Split(raw, "/") {
		if part == ".." {
			return "", fmt.Errorf("invalid path")
		}
	}
	clean := path.Clean("/" + raw)
	if _, err := filepath.Localize(strings.TrimPrefix(clean, "/")); clean != "/" && err != nil {
		return "", err
	}
	return clean, nil
}
func pageNumber(r *http.Request) (int, error) {
	p := r.URL.Query().Get("p")
	if p == "" {
		return 1, nil
	}
	n, e := strconv.Atoi(p)
	if e != nil || n < 1 {
		return 0, fmt.Errorf("invalid page number")
	}
	return n, nil
}
func contextURL(route, target string, r *http.Request, page int) string {
	q := url.Values{"path": {target}}
	for _, key := range []string{"type", "q", "ext"} {
		for _, v := range r.URL.Query()[key] {
			q.Add(key, v)
		}
	}
	if page > 1 {
		q.Set("p", strconv.Itoa(page))
	}
	return route + "?" + q.Encode()
}
func listing(hfs *fs.HeheFS, dir string, r *http.Request) ([]models.GalleryItem, error) {
	f, e := hfs.Open(dir)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	stat, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !stat.IsDir() {
		return nil, errNotDirectory
	}
	entries, e := f.Readdir(-1)
	if e != nil {
		return nil, e
	}
	filter := parseGalleryFilter(r)
	items := []models.GalleryItem{}
	for _, entry := range entries {
		it := models.GalleryItem{Filename: entry.Name(), Path: strings.TrimSuffix(dir, "/") + "/", IsDir: entry.IsDir(), Size: entry.Size(), ModTime: entry.ModTime()}
		if filter.Matches(&it) {
			items = append(items, it)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := strings.ToLower(items[i].Filename), strings.ToLower(items[j].Filename)
		if a == b {
			return items[i].Filename < items[j].Filename
		}
		return a < b
	})
	return items, nil
}
func serializeItem(it models.GalleryItem, cfg *config.Config, r *http.Request, page int) mediaItem {
	full := it.Path + it.Filename
	out := mediaItem{Name: it.Filename, Path: full, Type: it.TypeCategory(), Size: it.Size, Modified: it.ModTime, Width: it.Width, Height: it.Height, URL: it.GetUrl(), ViewerURL: contextURL("/post/", full, r, page)}
	if it.IsDir {
		out.URL = contextURL("/", full+"/", r, 1)
	} else if it.IsImage() {
		out.Thumbnail = out.URL
		if cfg.Resize && it.IsResizable() {
			out.Thumbnail = it.GetResized()
		}
	} else if it.IsVideo() && cfg.Resize && cfg.FFmpegExists {
		out.Thumbnail = it.GetVidThumb()
	}
	return out
}
func GalleryAPI(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, cfg *config.Config) {
	dir, e := apiPath(ctx)
	if e != nil {
		apiError(w, 400, e.Error())
		return
	}
	page, e := pageNumber(r)
	if e != nil {
		apiError(w, 400, e.Error())
		return
	}
	items, e := listing(hfs, dir, r)
	if e != nil {
		if errors.Is(e, errNotDirectory) {
			apiError(w, http.StatusBadRequest, e.Error())
			return
		}
		apiError(w, utils.StatusForErr(e), "Cannot read directory")
		return
	}
	pages := max(1, (len(items)+cfg.Split-1)/cfg.Split)
	if page > pages {
		apiError(w, 400, "Page out of range")
		return
	}
	start := min((page-1)*cfg.Split, len(items))
	batch := items[start:min(start+cfg.Split, len(items))]
	probeImageDimensions(hfs, batch)
	out := galleryResponse{Items: []mediaItem{}, Path: strings.TrimSuffix(dir, "/") + "/", Page: page, Pages: pages, Total: len(items), Infinite: !cfg.NoInfiniteScroll, Version: version.GetVersion()}
	for _, it := range batch {
		out.Items = append(out.Items, serializeItem(it, cfg, r, page))
	}
	jsonReply(w, 200, out)
}
func MediaAPI(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, cfg *config.Config) {
	full, e := apiPath(ctx)
	if e != nil {
		apiError(w, 400, e.Error())
		return
	}
	page, e := pageNumber(r)
	if e != nil {
		apiError(w, 400, e.Error())
		return
	}
	f, e := hfs.Open(full)
	if e != nil {
		apiError(w, utils.StatusForErr(e), "Cannot open media")
		return
	}
	stat, e := f.Stat()
	f.Close()
	if e != nil {
		apiError(w, utils.StatusForErr(e), "Cannot read media")
		return
	}
	if stat.IsDir() {
		apiError(w, 400, "Expected a file")
		return
	}
	dir := path.Dir(full)
	items, e := listing(hfs, dir, r)
	if e != nil {
		apiError(w, utils.StatusForErr(e), "Cannot read directory")
		return
	}
	current := models.GalleryItem{Filename: stat.Name(), Path: strings.TrimSuffix(dir, "/") + "/", Size: stat.Size(), ModTime: stat.ModTime()}
	out := mediaResponse{Item: serializeItem(current, cfg, r, page), GalleryURL: contextURL("/", strings.TrimSuffix(dir, "/")+"/", r, page)}
	media := []models.GalleryItem{}
	index := -1
	for _, it := range items {
		if !it.IsDir && (it.IsImage() || it.IsVideo() || it.IsAudio()) {
			if it.Filename == current.Filename {
				index = len(media)
			}
			media = append(media, it)
		}
	}
	out.Total = len(media)
	if index >= 0 {
		out.Position = index + 1
		if index > 0 {
			out.Previous = contextURL("/post/", media[index-1].Path+media[index-1].Filename, r, page)
		}
		if index+1 < len(media) {
			out.Next = contextURL("/post/", media[index+1].Path+media[index+1].Filename, r, page)
		}
	}
	jsonReply(w, 200, out)
}
