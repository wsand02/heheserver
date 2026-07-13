package templates

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSeq(t *testing.T) {
	cases := []struct {
		start, end int
		want       []int
	}{
		{1, 3, []int{1, 2, 3}},
		{3, 1, []int{}},
		{2, 2, []int{2}},
	}
	for _, c := range cases {
		if got := seq(c.start, c.end); !reflect.DeepEqual(got, c.want) {
			t.Errorf("seq(%d, %d) = %v, want %v", c.start, c.end, got, c.want)
		}
	}
}

func TestArithmeticHelpers(t *testing.T) {
	if sub(5, 3) != 2 {
		t.Errorf("sub(5,3) = %d, want 2", sub(5, 3))
	}
	if add(5, 3) != 8 {
		t.Errorf("add(5,3) = %d, want 8", add(5, 3))
	}
	if !gt(5, 3) || gt(3, 5) || gt(3, 3) {
		t.Error("gt behaves incorrectly")
	}
	if !lt(3, 5) || lt(5, 3) || lt(3, 3) {
		t.Error("lt behaves incorrectly")
	}
	if !ge(3, 3) || !ge(5, 3) || ge(3, 5) {
		t.Error("ge behaves incorrectly")
	}
	if !le(3, 3) || !le(3, 5) || le(5, 3) {
		t.Error("le behaves incorrectly")
	}
}

func TestRenderError(t *testing.T) {
	rec := httptest.NewRecorder()
	RenderError(rec, http.StatusNotFound, "open file")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{"404", http.StatusText(http.StatusNotFound), "open file", "Back to gallery"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\n%s", want, body)
		}
	}
}

func TestStaticHandler(t *testing.T) {
	ts := httptest.NewServer(StaticHandler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/static/glacialwisp.min.css")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty CSS body")
	}
}
