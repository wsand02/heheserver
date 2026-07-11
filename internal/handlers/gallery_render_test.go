package handlers

import (
	"html"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wsand02/heheserver/internal/models"
	"github.com/wsand02/heheserver/internal/templates"
)

// renderList renders the "list" template with the given context and returns the body.
func renderList(t *testing.T, gc *GalleryContext) string {
	t.Helper()
	w := httptest.NewRecorder()
	templates.RenderTemplate(w, "list", gc)
	if w.Code != 200 {
		t.Fatalf("render list: status %d, body: %s", w.Code, w.Body.String())
	}
	return w.Body.String()
}

func renderPost(t *testing.T, pc *PostContext) string {
	t.Helper()
	w := httptest.NewRecorder()
	templates.RenderTemplate(w, "post", pc)
	if w.Code != 200 {
		t.Fatalf("render post: status %d, body: %s", w.Code, w.Body.String())
	}
	return w.Body.String()
}

// countBreadcrumbs returns how many breadcrumb <li> items the rendered body has.
func countBreadcrumbs(body string) int {
	return strings.Count(body, `class="breadcrumb-item"`)
}

// TestListRootBreadcrumbsNoDuplicates guards against the split("/", "/") == ["", ""]
// bug that rendered "Home" plus two empty breadcrumb items on the root page.
func TestListRootBreadcrumbsNoDuplicates(t *testing.T) {
	gc := &GalleryContext{Path: "/", CurrentPage: 1, MaxPage: 1}
	if n := countBreadcrumbs(renderList(t, gc)); n != 1 {
		t.Errorf("root page: expected 1 breadcrumb (Home), got %d", n)
	}
}

// TestListNestedBreadcrumbs checks a nested path yields exactly Home + one crumb
// per path segment, with no stray empty leading crumb.
func TestListNestedBreadcrumbs(t *testing.T) {
	gc := &GalleryContext{Path: "/foo/bar/", CurrentPage: 1, MaxPage: 1}
	if n := countBreadcrumbs(renderList(t, gc)); n != 3 {
		t.Errorf("/foo/bar/: expected 3 breadcrumbs (Home, foo, bar), got %d", n)
	}
}

// TestListHasResultsWrapper guards the realtime-filter swap target: the grid,
// status, and pagination must live inside #gallery-results (and the filter form
// must stay outside it, so its focused input survives a swap).
func TestListHasResultsWrapper(t *testing.T) {
	gc := &GalleryContext{Path: "/", CurrentPage: 1, MaxPage: 1}
	body := renderList(t, gc)

	if !strings.Contains(body, `id="gallery-results"`) {
		t.Fatal("missing #gallery-results swap container")
	}
	formIdx := strings.Index(body, `class="filter-bar"`)
	resultsIdx := strings.Index(body, `id="gallery-results"`)
	gridIdx := strings.Index(body, `class="grid"`)
	if formIdx == -1 || formIdx > resultsIdx {
		t.Error("filter form must render before (outside) #gallery-results")
	}
	if gridIdx < resultsIdx {
		t.Error("grid must render inside #gallery-results")
	}
}

// TestListEscapesFilename is the reason for the html/template switch: a
// user-controlled filename containing markup must be rendered as inert text, not
// injected as live HTML.
func TestListEscapesFilename(t *testing.T) {
	gc := &GalleryContext{
		Path:        "/",
		CurrentPage: 1,
		MaxPage:     1,
		Items: []models.GalleryItem{
			{Filename: "<script>alert(1)</script>.txt", Path: "/"},
		},
	}
	body := renderList(t, gc)

	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Errorf("filename rendered as live HTML (XSS): body contains raw <script>")
	}
	if !strings.Contains(body, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("expected HTML-escaped filename in output, got:\n%s", body)
	}
}

// TestListSpecialCharURLsNotBroken guards the CLAUDE.md special-char 404 class:
// filenames with spaces/emoji must still produce correctly percent-encoded URLs
// and must not be double-encoded or filtered to #ZgotmplZ.
func TestListSpecialCharURLsNotBroken(t *testing.T) {
	gc := &GalleryContext{
		Path:        "/sub dir/",
		CurrentPage: 1,
		MaxPage:     1,
		Items: []models.GalleryItem{
			{Filename: "a b 😀.txt", Path: "/sub dir/"},
		},
	}
	// html/template HTML-encodes some URL chars in attributes (e.g. + -> &#43;);
	// the browser decodes those before using the URL, so assert against the
	// effective (HTML-unescaped) URLs — this is what actually reaches the server.
	body := html.UnescapeString(renderList(t, gc))

	if strings.Contains(body, "ZgotmplZ") {
		t.Errorf("a URL was filtered to #ZgotmplZ:\n%s", body)
	}
	// /fs/ path URL from GetUrl: space -> %20, emoji percent-encoded, no double %25.
	if !strings.Contains(body, "/fs/sub%20dir/a%20b%20") {
		t.Errorf("expected percent-encoded /fs/ URL, got:\n%s", body)
	}
	// ?path= post link from GetPostLink: QueryEscape encodes space -> +.
	if !strings.Contains(body, "?path=/sub+dir/a+b+") {
		t.Errorf("expected query-escaped post link, got:\n%s", body)
	}
	if strings.Contains(body, "%25") {
		t.Errorf("URL appears double-encoded (contains %%25):\n%s", body)
	}
}

// TestPostSpecialCharURLsNotBroken does the same guard for the single-item view.
func TestPostSpecialCharURLsNotBroken(t *testing.T) {
	pc := &PostContext{models.GalleryItem{Filename: "a b 😀.txt", Path: "/sub dir/"}}
	body := html.UnescapeString(renderPost(t, pc))

	if strings.Contains(body, "ZgotmplZ") {
		t.Errorf("a URL was filtered to #ZgotmplZ:\n%s", body)
	}
	if !strings.Contains(body, "/fs/sub%20dir/") {
		t.Errorf("expected percent-encoded /fs/ path URL, got:\n%s", body)
	}
	if strings.Contains(body, "%25") {
		t.Errorf("URL appears double-encoded (contains %%25):\n%s", body)
	}
}
