package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/models"
	"github.com/wsand02/heheserver/internal/templates"
	"github.com/wsand02/heheserver/internal/utils"
)

type GalleryContext struct {
	Items        []models.GalleryItem
	Resize       bool
	FFmpegExists bool
	Path         string
	CurrentPage  int
	MaxPage      int
	Filter       models.GalleryFilter
	FilterQuery  string // raw ?q= echo, to refill the search input
	FilterExt    string // raw ?ext= echo, to refill the extension input
}

// typeOrder is the stable rendering/serialization order for file-type
// categories (checkboxes, data-type attribute, ?type= params).
var typeOrder = []string{"image", "video", "audio", "dir", "other"}

// TypeOptions returns the file-type categories offered in the filter bar.
func (gc *GalleryContext) TypeOptions() []string {
	return typeOrder
}

// TypeChecked reports whether file-type category t is active, for rendering
// checkbox checked state in the filter bar.
func (gc *GalleryContext) TypeChecked(t string) bool {
	return gc.Filter.Types[t]
}

// ClearURL is the current directory with all filters removed (the "Clear" link).
func (gc *GalleryContext) ClearURL() string {
	q := url.Values{}
	if gc.Path != "" {
		q.Set("path", gc.Path)
	}
	if len(q) == 0 {
		return "?"
	}
	return "?" + q.Encode()
}

// activeTypes returns the active file-type categories in stable order.
func (gc *GalleryContext) activeTypes() []string {
	if len(gc.Filter.Types) == 0 {
		return nil
	}
	var out []string
	for _, t := range typeOrder {
		if gc.Filter.Types[t] {
			out = append(out, t)
		}
	}
	return out
}

// TypesParam joins the active file-type categories with commas for the grid's
// data-type attribute, which infinite scroll reads to carry the filter forward.
func (gc *GalleryContext) TypesParam() string {
	return strings.Join(gc.activeTypes(), ",")
}

// PageURL builds the gallery URL for the given page number, preserving the
// active filter params so the no-JS pagination links keep the filter applied.
func (gc *GalleryContext) PageURL(page int) string {
	q := url.Values{}
	if gc.Path != "" {
		q.Set("path", gc.Path)
	}
	if page > 1 { // omit p=1 for a cleaner URL
		q.Set("p", strconv.Itoa(page))
	}
	for _, t := range gc.activeTypes() {
		q.Add("type", t)
	}
	if gc.FilterQuery != "" {
		q.Set("q", gc.FilterQuery)
	}
	if gc.FilterExt != "" {
		q.Set("ext", gc.FilterExt)
	}
	if len(q) == 0 {
		return "?"
	}
	return "?" + q.Encode()
}

// parseGalleryFilter reads the filter params (?type=, ?q=, ?ext=) from the
// request into a models.GalleryFilter. Unknown type values are ignored;
// extensions are normalized to a lowercased leading-dot form.
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

// GetBreadcrumbs returns the non-empty path segments. Splitting on "/" yields
// empty leading/trailing segments (e.g. "/" -> ["", ""], "/foo/bar/" ->
// ["", "foo", "bar", ""]); those would render as blank breadcrumb items, so
// they are dropped here. The hardcoded "Home" crumb in the template covers the
// root, so an empty slice is correct for "/".
func (gc *GalleryContext) GetBreadcrumbs() []string {
	var parts []string
	for _, p := range strings.Split(gc.Path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func (gc *GalleryContext) BreadcrumbToUrl(i int) string {
	crumbs := gc.GetBreadcrumbs()[: i+1]
	pcrumbs := []string{"?path="}
	for _, c := range crumbs {
		pcrumbs = append(pcrumbs, url.QueryEscape(c))
	}
	u := strings.Join(pcrumbs, "/")
	urla := strings.Join([]string{u, "/"}, "")
	return urla
}

func GalleryHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		utils.HttpLogErr(w, r, err, "open file", utils.StatusForErr(err))
		return
	}
	defer hf.Close()
	dirlis, err := hf.Readdir(-1)
	if err != nil {
		utils.HttpLogErr(w, r, err, "read directory", http.StatusInternalServerError)
		return
	}

	filter := parseGalleryFilter(r)

	// Build the gallery items, applying the filter before pagination so the
	// whole directory is filtered rather than just the pages already loaded.
	var items []models.GalleryItem
	for _, item := range dirlis {
		gi := models.GalleryItem{Filename: item.Name(), IsDir: item.IsDir(), Size: item.Size(), ModTime: item.ModTime(), Path: ctx}
		if filter.Active() && !filter.Matches(&gi) {
			continue
		}
		items = append(items, gi)
	}

	// Paginate the (filtered) items.
	var divided [][]models.GalleryItem
	for i := 0; i < len(items); i += config.Split {
		divided = append(divided, items[i:min(i+config.Split, len(items))])
	}

	q := r.URL.Query().Get("p")
	if q == "" {
		q = "1"
	}
	pid, err := strconv.Atoi(q)
	if err != nil {
		utils.HttpLogErr(w, r, err, "invalid page number", http.StatusBadRequest)
		return
	}
	pid -= 1

	gc := GalleryContext{
		Resize:       config.Resize,
		FFmpegExists: config.FFmpegExists,
		Path:         ctx,
		CurrentPage:  pid + 1,
		MaxPage:      len(divided),
		Filter:       filter,
		FilterQuery:  strings.TrimSpace(r.URL.Query().Get("q")),
		FilterExt:    strings.TrimSpace(r.URL.Query().Get("ext")),
	}

	if pid < 0 || pid >= len(divided) {
		// An empty result (empty directory, or a filter that matched nothing)
		// is a valid page 1 with zero items — never fall back to the unfiltered
		// listing. Any other out-of-range page is a bad request.
		if pid == 0 && len(divided) == 0 {
			gc.CurrentPage = 1
			gc.MaxPage = 1
			templates.RenderTemplate(w, "list", &gc)
			return
		}
		utils.HttpLogErr(w, r, fmt.Errorf("p: %d out of range", pid), "page out of range", http.StatusBadRequest)
		return
	}

	gc.Items = divided[pid]
	templates.RenderTemplate(w, "list", &gc)
}

type PostContext struct {
	models.GalleryItem
}

// GalleryURL is the listing URL for the directory containing this item. It backs
// the "Back to gallery" link's no-JS fallback (with JS, the link prefers
// history.back() so the exact filtered/scrolled gallery state is restored).
// Here Path is the item's full path (e.g. "/folder/file.png"), so we strip the
// trailing filename to get the directory.
func (pc *PostContext) GalleryURL() string {
	dir := "/"
	if i := strings.LastIndex(strings.TrimSuffix(pc.Path, "/"), "/"); i >= 0 {
		dir = pc.Path[:i+1]
	}
	q := url.Values{}
	q.Set("path", dir)
	return "?" + q.Encode()
}

func PostHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		utils.HttpLogErr(w, r, err, "open file", utils.StatusForErr(err))
		return
	}
	defer hf.Close()
	hfstat, err := hf.Stat()
	if err != nil {
		utils.HttpLogErr(w, r, err, "stat file", utils.StatusForErr(err))
		return
	}
	if hfstat.IsDir() {
		utils.HttpLogErr(w, r, fmt.Errorf("directory on posthandler"), "this is a directory", http.StatusBadRequest)
		return
	}
	templates.RenderTemplate(w, "post", &PostContext{models.GalleryItem{Filename: hfstat.Name(), IsDir: hfstat.IsDir(), Path: ctx, Size: hfstat.Size(), ModTime: hfstat.ModTime()}}) // oh well
}
