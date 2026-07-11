package vidthumb

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/utils"
)

// makeTestVideo generates a short test clip with ffmpeg and returns its path.
func makeTestVideo(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "clip.mp4")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", "testsrc=duration=2:size=64x64:rate=10",
		"-pix_fmt", "yuv420p", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to generate test video: %v\n%s", err, out)
	}
	return path
}

func TestGenerateThumb(t *testing.T) {
	if !utils.FFmpegExists() {
		t.Skip("ffmpeg not on PATH")
	}
	img, err := GenerateThumb(makeTestVideo(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Fatal("expected a thumbnail image, got nil")
	}
	if img.Bounds().Empty() {
		t.Error("thumbnail has empty bounds")
	}
}
