package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"syscall"
	"testing"
)

func TestIsClientDisconnect(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"connection closed", net.ErrClosed, true},
		{"connection reset", fmt.Errorf("write tcp: %w", syscall.ECONNRESET), true},
		{"connection aborted", fmt.Errorf("write tcp: %w", syscall.ECONNABORTED), true},
		{"broken pipe", syscall.EPIPE, true},
		{"other error", errors.New("encoding failed"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isClientDisconnect(c.err); got != c.want {
				t.Errorf("isClientDisconnect(%v) = %t, want %t", c.err, got, c.want)
			}
		})
	}
}

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
