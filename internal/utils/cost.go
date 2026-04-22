package utils

import "image"

func GetCost(img image.Image) int64 {
	return int64(img.Bounds().Dx() * img.Bounds().Dy())
}
