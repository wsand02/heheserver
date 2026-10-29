package handlers

import (
	"encoding/json"
	"image"
	"image/color"
	"image/gif"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
)

func TestAnimatedGIFGalleryAndViewer(t *testing.T) {
	dir := t.TempDir()
	palette := color.Palette{color.Black, color.White}
	frames := []*image.Paletted{image.NewPaletted(image.Rect(0, 0, 2, 3), palette), image.NewPaletted(image.Rect(0, 0, 2, 3), palette)}
	frames[1].SetColorIndex(0, 0, 1)
	f, err := os.Create(filepath.Join(dir, "animated.GIF"))
	if err != nil {
		t.Fatal(err)
	}
	err = gif.EncodeAll(f, &gif.GIF{Image: frames, Delay: []int{10, 10}})
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	writePNG(t, filepath.Join(dir, "next.png"), 2, 3)
	hfs := testHFS(t, dir)
	cfg := &config.Config{Split: 1, Resize: true}
	w := httptest.NewRecorder()
	GalleryAPI(w, httptest.NewRequest("GET", "/api/gallery?type=image&ext=gif", nil), "/", hfs, cfg)
	var gallery galleryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &gallery); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || len(gallery.Items) != 1 {
		t.Fatalf("GIF listing: %d %s", w.Code, w.Body.String())
	}
	item := gallery.Items[0]
	if item.Type != "image" || item.Width != 2 || item.Height != 3 || item.Thumbnail != item.URL || item.URL != "/fs/animated.GIF" {
		t.Fatalf("GIF must retain its original animated preview: %+v", item)
	}
	w = httptest.NewRecorder()
	MediaAPI(w, httptest.NewRequest("GET", "/api/media?type=image", nil), "/animated.GIF", hfs, cfg)
	var media mediaResponse
	if err := json.Unmarshal(w.Body.Bytes(), &media); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || media.Item.Type != "image" || media.Position != 1 || media.Total != 2 || media.Next == "" {
		t.Fatalf("GIF viewer navigation: %d %+v", w.Code, media)
	}
}

func testHFS(t *testing.T, dir string) *fs.HeheFS {
	t.Helper()
	if err := cache.NewIgnoreCache(1); err != nil {
		t.Fatal(err)
	}
	if err := cache.NewDimensionCache(); err != nil {
		t.Fatal(err)
	}
	return fs.Dir(dir).(*fs.HeheFS)
}
func TestGalleryAndMediaAPI(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"A.png", "b.mp4", "c++.png", "note.txt", "secret.png"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("secret.png\n.heheignore\n"), 0600); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cfg := &config.Config{Split: 1, NoInfiniteScroll: true}
	request := func(endpoint, target string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("GET", target, nil)
		w := httptest.NewRecorder()
		ctx := r.URL.Query().Get("path")
		if endpoint == "gallery" {
			GalleryAPI(w, r, ctx, hfs, cfg)
		} else {
			MediaAPI(w, r, ctx, hfs, cfg)
		}
		return w
	}
	w := request("gallery", "/api/gallery?path=/&type=image&p=2")
	var g galleryResponse
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &g); err != nil {
		t.Fatal(err)
	}
	if g.Total != 2 || g.Pages != 2 || g.Infinite || len(g.Items) != 1 || g.Items[0].Name != "c++.png" {
		t.Fatalf("unexpected gallery: %+v", g)
	}
	u, err := url.Parse(g.Items[0].ViewerURL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("path") != "/c++.png" || u.Query().Get("p") != "2" {
		t.Fatal(u)
	}
	w = request("media", "/api/media"+u.String()[len("/post/"):])
	var m mediaResponse
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m.Position != 2 || m.Total != 2 || m.Previous == "" || m.Next != "" {
		t.Fatalf("unexpected media: %+v", m)
	}
	back, _ := url.Parse(m.GalleryURL)
	if back.Query().Get("type") != "image" || back.Query().Get("p") != "2" {
		t.Fatal(back)
	}
	w = request("media", "/api/media?path=/b.mp4")
	json.Unmarshal(w.Body.Bytes(), &m)
	if m.Position != 2 || m.Total != 3 || m.Previous == "" || m.Next == "" {
		t.Fatalf("cross-page neighbors: %+v", m)
	}
	for _, target := range []string{"/api/gallery?p=0", "/api/gallery?p=abc", "/api/gallery?p=999", "/api/gallery?path=/../secret", "/api/gallery?path=/a%5Cb", "/api/gallery?path=/note.txt"} {
		if w := request("gallery", target); w.Code != 400 {
			t.Errorf("%s: %d", target, w.Code)
		}
	}
	for _, target := range []string{"/api/media?path=/secret.png", "/api/media?path=/missing.png"} {
		w := request("media", target)
		if w.Code != 404 || w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("%s: %d %s", target, w.Code, w.Body.String())
		}
	}
	if w := request("media", "/api/media?path=/folder"); w.Code != 400 {
		t.Fatal(w.Code)
	}
	w = request("gallery", "/api/gallery?q=nothing")
	json.Unmarshal(w.Body.Bytes(), &g)
	if len(g.Items) != 0 || g.Pages != 1 {
		t.Fatalf("empty results: %+v", g)
	}
	if err := os.Remove(filepath.Join(dir, "b.mp4")); err != nil {
		t.Fatal(err)
	}
	if w := request("media", "/api/media?path=/b.mp4"); w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestAPIFiltersCombineBeforePagination(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"Cat.jpg", "cat.png", "dog.jpg", "song.mp3"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	hfs := testHFS(t, dir)
	for _, tc := range []struct {
		query string
		total int
		first string
	}{
		{"type=image&q=CAT&ext=JPG", 1, "Cat.jpg"},
		{"type=image&type=audio&ext=png,mp3", 2, "cat.png"},
		{"type=unknown&q=dog", 1, "dog.jpg"},
		{"type=video", 0, ""},
	} {
		r := httptest.NewRequest("GET", "/api/gallery?"+tc.query, nil)
		w := httptest.NewRecorder()
		GalleryAPI(w, r, "/", hfs, &config.Config{Split: 1})
		var got galleryResponse
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || got.Total != tc.total {
			t.Fatalf("%s: %d %+v", tc.query, w.Code, got)
		}
		if tc.first != "" && (len(got.Items) != 1 || got.Items[0].Name != tc.first) {
			t.Fatalf("%s: %+v", tc.query, got.Items)
		}
	}
}

func TestAPISpecialPathsAndBoundaries(t *testing.T) {
	dir := t.TempDir()
	hfs := testHFS(t, dir)
	cfg := &config.Config{Split: 2}
	for _, name := range []string{"é & #.png", "😀.mp4", "song.ogg"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"é & #.png", "😀.mp4", "song.ogg"} {
		r := httptest.NewRequest(http.MethodGet, "/api/media?"+url.Values{"path": {"/" + name}, "q": {name}}.Encode(), nil)
		w := httptest.NewRecorder()
		MediaAPI(w, r, "/"+name, hfs, cfg)
		var m mediaResponse
		json.Unmarshal(w.Body.Bytes(), &m)
		if w.Code != 200 || m.Position != 1 || m.Total != 1 || m.Next != "" || m.Previous != "" {
			t.Fatalf("%s: %d %+v", name, w.Code, m)
		}
		u, _ := url.Parse(m.Item.URL)
		f, err := hfs.Open(u.Path[len("/fs"):])
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}
