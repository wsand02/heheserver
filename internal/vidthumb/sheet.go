package vidthumb

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var durationPattern = regexp.MustCompile(`Duration: (\d+):(\d+):(\d+(?:\.\d+)?)`)

// GeneratePreview falls back to the original thumbnail if a sheet cannot be made.
func GeneratePreview(path string, sheet bool) (image.Image, error) {
	if sheet {
		return previewWithSheet(path, GenerateSheet)
	}
	return GenerateThumb(path)
}

func previewWithSheet(path string, generate func(string) (image.Image, error)) (image.Image, error) {
	if img, err := generate(path); err == nil {
		return img, nil
	}
	return GenerateThumb(path)
}

// GenerateSheet samples four timestamps under one shared processing deadline.
func GenerateSheet(path string) (image.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return generateSheet(ctx, path)
}

func generateSheet(ctx context.Context, path string) (image.Image, error) {
	// ffmpeg reports input metadata on stderr even though no output is specified.
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-nostdin", "-i", path)
	var metadata bytes.Buffer
	cmd.Stderr = &metadata
	_ = cmd.Run()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	match := durationPattern.FindStringSubmatch(metadata.String())
	if match == nil {
		return nil, fmt.Errorf("video duration unavailable")
	}
	hours, _ := strconv.ParseFloat(match[1], 64)
	minutes, _ := strconv.ParseFloat(match[2], 64)
	seconds, _ := strconv.ParseFloat(match[3], 64)
	duration := hours*3600 + minutes*60 + seconds
	if duration <= 0 {
		return nil, fmt.Errorf("video duration must be positive")
	}
	sheet := image.NewRGBA(image.Rect(0, 0, 640, 360))
	for i, fraction := range []float64{0.10, 0.35, 0.60, 0.85} {
		target := duration * fraction
		// Decode a short lead-in and extend the last frame so samples between
		// the final frame timestamp and container duration still produce an image.
		seek := max(0, target-1)
		cmd := exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-v", "error", "-ss", strconv.FormatFloat(seek, 'f', 6, 64), "-i", path,
			"-ss", strconv.FormatFloat(target-seek, 'f', 6, 64),
			"-map", "0:v:0", "-frames:v", "1", "-vf", "scale=320:180:force_original_aspect_ratio=decrease,pad=320:180:(ow-iw)/2:(oh-ih)/2:black,setsar=1,tpad=stop_mode=clone:stop_duration=1",
			"-f", "image2pipe", "-vcodec", "png", "-")
		var frame bytes.Buffer
		cmd.Stdout = &frame
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("extract sheet frame: %w", err)
		}
		img, _, err := image.Decode(&frame)
		if err != nil {
			return nil, fmt.Errorf("decode sheet frame: %w", err)
		}
		cell := image.Rect((i%2)*320, (i/2)*180, (i%2+1)*320, (i/2+1)*180)
		draw.Draw(sheet, cell, img, img.Bounds().Min, draw.Src)
	}
	return sheet, nil
}
