package vidthumb

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
	args := []string{"-nostdin", "-v", "error"}
	var filters []string
	for i, fraction := range []float64{0.10, 0.35, 0.60, 0.85} {
		target := duration * fraction
		// Each input seeks near its sample; only the short lead-in is decoded.
		// Extend the last frame for samples beyond the final frame timestamp.
		seek := max(0, target-1)
		args = append(args, "-ss", strconv.FormatFloat(seek, 'f', 6, 64), "-i", path)
		filters = append(filters, fmt.Sprintf(
			"[%d:v:0]scale=320:180:force_original_aspect_ratio=decrease,pad=320:180:(ow-iw)/2:(oh-ih)/2:black,setsar=1,tpad=stop_mode=clone:stop_duration=1,trim=start=%.6f,trim=end_frame=1,setpts=PTS-STARTPTS[cell%d]",
			i, target-seek, i))
	}
	filters = append(filters, "[cell0][cell1][cell2][cell3]xstack=inputs=4:layout=0_0|320_0|0_180|320_180[sheet]")
	// Return one raw frame, avoiding four PNG encodes/decodes and Go-side
	// composition. The handler performs the final JPEG encode as before.
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[sheet]",
		"-frames:v", "1", "-f", "rawvideo", "-pix_fmt", "rgba", "-")
	cmd = exec.CommandContext(ctx, "ffmpeg", args...)
	var frame, stderr bytes.Buffer
	cmd.Stdout = &frame
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("extract sheet: %w: %s", err, stderr.String())
	}
	sheet := image.NewRGBA(image.Rect(0, 0, 640, 360))
	if frame.Len() != len(sheet.Pix) {
		return nil, fmt.Errorf("unexpected sheet size: got %d bytes, want %d", frame.Len(), len(sheet.Pix))
	}
	copy(sheet.Pix, frame.Bytes())
	return sheet, nil
}
