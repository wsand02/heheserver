package vidthumb

import (
	"context"
	"errors"
	"image"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/utils"
)

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if !utils.FFmpegExists() {
		t.Skip("ffmpeg not on PATH")
	}
}

func TestGenerateSheetFrameOrder(t *testing.T) {
	requireFFmpeg(t)
	path := filepath.Join(t.TempDir(), "colors.mp4")
	args := []string{"-y"}
	for _, color := range []string{"red", "lime", "blue", "white"} {
		args = append(args, "-f", "lavfi", "-i", "color=c="+color+":s=64x64:r=10:d=1")
	}
	args = append(args, "-filter_complex", "[0:v][1:v][2:v][3:v]concat=n=4:v=1:a=0[v]", "-map", "[v]", "-c:v", "mpeg4", path)
	if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v: %s", err, out)
	}
	img, err := GenerateSheet(path)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds() != image.Rect(0, 0, 640, 360) {
		t.Fatal(img.Bounds())
	}
	for i, want := range [][3]uint32{{255, 0, 0}, {0, 255, 0}, {0, 0, 255}, {255, 255, 255}} {
		r, g, b, _ := img.At(i%2*320+160, i/2*180+90).RGBA()
		for channel, got := range []uint32{r >> 8, g >> 8, b >> 8} {
			diff := int(got) - int(want[channel])
			if diff < -10 || diff > 10 {
				t.Fatalf("cell %d: RGB=%v want=%v", i, []uint32{r >> 8, g >> 8, b >> 8}, want)
			}
		}
		r, g, b, _ = img.At(i%2*320+5, i/2*180+90).RGBA()
		if r|g|b != 0 {
			t.Fatalf("cell %d: padding is not black", i)
		}
	}
}

func TestGenerateSheetShortVideo(t *testing.T) {
	requireFFmpeg(t)
	path := filepath.Join(t.TempDir(), "short.mp4")
	if out, err := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "color=c=red:s=64x64:r=30:d=0.2", "-c:v", "mpeg4", path).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v: %s", err, out)
	}
	if _, err := GenerateSheet(path); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateSheetInvalidAndCanceled(t *testing.T) {
	requireFFmpeg(t)
	if _, err := GeneratePreview(filepath.Join(t.TempDir(), "missing.mp4"), true); err == nil {
		t.Fatal("expected invalid video error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := generateSheet(ctx, "unused"); err != context.Canceled {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestSheetFailureFallsBackToThumbnail(t *testing.T) {
	requireFFmpeg(t)
	img, err := previewWithSheet(makeTestVideo(t), func(string) (image.Image, error) {
		return nil, errors.New("sheet failed")
	})
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds() != image.Rect(0, 0, 64, 64) {
		t.Fatalf("expected original thumbnail dimensions, got %v", img.Bounds())
	}
}
