package resize

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
)

func ResizePng(w io.Writer, in io.Reader) error {
	src, err := png.Decode(in)
	if err != nil {
		return err
	}
	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	err = png.Encode(w, dst)
	if err != nil {
		return err
	}
	return nil
}

func ResizeJpeg(w io.Writer, in io.Reader) error {
	src, err := jpeg.Decode(in)
	if err != nil {
		return err
	}
	dst := image.NewRGBA(image.Rect(0, 0, src.Bounds().Max.X/2, src.Bounds().Max.Y/2))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	err = jpeg.Encode(w, dst, nil)
	if err != nil {
		return err
	}
	return nil
}
