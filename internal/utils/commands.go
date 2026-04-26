package utils

import "os/exec"

func FFmpegExists() bool {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
