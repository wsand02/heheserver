package utils

import (
	"fmt"
	"io/fs"
	"net/http"
	"testing"
)

func TestStatusForErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"not exist", fmt.Errorf("open x: %w", fs.ErrNotExist), http.StatusNotFound},
		{"permission", fmt.Errorf("open x: %w", fs.ErrPermission), http.StatusForbidden},
		{"generic", fmt.Errorf("boom"), http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := StatusForErr(c.err); got != c.want {
				t.Errorf("StatusForErr(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}
