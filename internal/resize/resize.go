package resize

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"os/exec"

	"golang.org/x/image/draw"
)

const width int = 300

func ResizeImage(path string) (image.Image, error) {
	cmd := exec.Command("ffmpeg", "-i", path, "-vf", fmt.Sprintf("scale=%d:-2", width), "-f", "image2pipe", "-")

	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(&out)
	if err != nil {
		log.Fatal(err)
	}
	return img, nil
}

func ResizeImageFallback(r io.Reader) (*image.RGBA, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return nil, fmt.Errorf("Image bounds x or y is 0")
	}
	ratio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	tgtH := int(float64(width) / ratio)
	dst := image.NewRGBA(image.Rect(0, 0, width, tgtH))
	draw.ApproxBiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	return dst, nil
}
