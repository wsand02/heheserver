package vidthumb

import (
	"bytes"
	"image"
	"log"
	"os/exec"
)

func GenerateThumb(path string) (image.Image, error) {
	cmd := exec.Command("ffmpeg", "-i", path, "-ss", "00:00:01.000", "-vframes", "1", "-f", "image2pipe", "-")

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
