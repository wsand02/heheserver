package resize

import (
	"image"

	"golang.org/x/image/draw"
)

const width int = 250

func ResizeImage(img image.Image) *image.RGBA {
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return nil
	}
	ratio := float64(img.Bounds().Dx()) / float64(img.Bounds().Dy())
	targetHeight := int(float64(width) / ratio)
	dst := image.NewRGBA(image.Rect(0, 0, width, targetHeight))
	draw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	return dst
}
