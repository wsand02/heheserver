package version

import (
	"runtime/debug"
	"testing"
)

func TestVersionFrom(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			// Regression: ReadBuildInfo returning ok == false yields a nil
			// *BuildInfo. Dereferencing it used to panic.
			name: "unavailable build info",
			info: nil,
			ok:   false,
			want: unknownVersion,
		},
		{
			name: "empty version",
			info: &debug.BuildInfo{Main: debug.Module{Version: ""}},
			ok:   true,
			want: unknownVersion,
		},
		{
			name: "tagged version",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			ok:   true,
			want: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionFrom(tt.info, tt.ok); got != tt.want {
				t.Errorf("versionFrom(%v, %v) = %q, want %q", tt.info, tt.ok, got, tt.want)
			}
		})
	}
}

func TestGetVersionDoesNotPanic(t *testing.T) {
	// In the test binary ReadBuildInfo generally succeeds; this just guards
	// that GetVersion never panics and always returns something.
	if got := GetVersion(); got == "" {
		t.Error("GetVersion() returned empty string")
	}
}
