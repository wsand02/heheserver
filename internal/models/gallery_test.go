package models

import "testing"

func TestGalleryItem_SizeMB(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		size int64
		want float64
	}{
		{
			name: "1MB",
			size: 1_000_000,
			want: 1.0,
		},
		{
			name: "0",
			size: 0,
			want: 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Size: tt.size}
			got := gi.SizeMB()
			if got != tt.want {
				t.Errorf("SizeMB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_IsImage(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		filename string
		want     bool
	}{
		{
			name:     "jpeg",
			filename: "hello.jpeg",
			want:     true,
		},
		{
			name:     "jpg",
			filename: "hello.jpg",
			want:     true,
		},
		{
			name:     "JPG",
			filename: "HELLO.JPG",
			want:     true,
		},
		{
			name:     "not image",
			filename: "notimage.exe",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Filename: tt.filename}
			got := gi.IsImage()
			if got != tt.want {
				t.Errorf("IsImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_IsVideo(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		filename string
		want     bool
	}{
		{
			name:     "movie",
			filename: "movie.mp4",
			want:     true,
		},
		{
			name:     "not movie",
			filename: "movie.exe",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Filename: tt.filename}
			got := gi.IsVideo()
			if got != tt.want {
				t.Errorf("IsVideo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_IsAudio(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		filename string
		want     bool
	}{
		{
			name:     "music",
			filename: "music.ogg",
			want:     true,
		},
		{
			name:     "not music",
			filename: "slop.exe",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Filename: tt.filename}
			got := gi.IsAudio()
			if got != tt.want {
				t.Errorf("IsAudio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_IsResizable(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		filename string
		want     bool
	}{
		{
			name:     "resizable",
			filename: "image.jpg",
			want:     true,
		},
		{
			name:     "not resizable",
			filename: "image.mp4",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Filename: tt.filename}
			got := gi.IsResizable()
			if got != tt.want {
				t.Errorf("IsResizable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_GetUrl(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		path     string
		filename string
		isDir    bool
		want     string
	}{
		{
			name:     "file",
			path:     "/folder/",
			filename: "file.file",
			isDir:    false,
			want:     "/fs/folder/file.file",
		},
		{
			name:     "folder",
			filename: "lefolder",
			path:     "/folder/",
			isDir:    true,
			want:     "?path=/folder/lefolder/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Filename: tt.filename, Path: tt.path, IsDir: tt.isDir}
			got := gi.GetUrl()
			if got != tt.want {
				t.Errorf("GetUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_GetPath(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		path string
		want string
	}{
		{
			name: "Path",
			path: "/hello/",
			want: "/fs/hello/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Path: tt.path}
			got := gi.GetPath()
			if got != tt.want {
				t.Errorf("GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_GetPostLink(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		path     string
		filename string
		want     string
	}{
		{
			name:     "PostLink",
			path:     "/hellothere/",
			filename: "hellosir.jpeg",
			want:     "/post/?path=/hellothere/hellosir.jpeg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Path: tt.path, Filename: tt.filename}
			got := gi.GetPostLink()
			if got != tt.want {
				t.Errorf("GetPostLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_GetResized(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		path     string
		filename string
		want     string
	}{
		{
			name:     "Resize",
			path:     "/hellothere/",
			filename: "hellosir.jpeg",
			want:     "/resize/?path=/hellothere/hellosir.jpeg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Path: tt.path, Filename: tt.filename}
			got := gi.GetResized()
			if got != tt.want {
				t.Errorf("GetResized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryItem_GetVidThumb(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		path     string
		filename string
		want     string
	}{
		{
			name:     "VidThumb",
			path:     "/hellothere/",
			filename: "hellosir.mp4",
			want:     "/vidthumb/?path=/hellothere/hellosir.mp4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Path: tt.path, Filename: tt.filename}
			got := gi.GetVidThumb()
			if got != tt.want {
				t.Errorf("GetVidThumb() = %v, want %v", got, tt.want)
			}
		})
	}
}
