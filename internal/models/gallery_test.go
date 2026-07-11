package models

import "testing"

func TestGalleryItem_TypeCategory(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		isDir    bool
		want     string
	}{
		{"dir", "photos", true, "dir"},
		{"image", "cat.png", false, "image"},
		{"video", "clip.mp4", false, "video"},
		{"audio", "song.ogg", false, "audio"},
		{"other", "notes.txt", false, "other"},
		// a directory named like an image is still a dir
		{"dir with image ext", "album.png", true, "dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := GalleryItem{Filename: tt.filename, IsDir: tt.isDir}
			if got := gi.TypeCategory(); got != tt.want {
				t.Errorf("TypeCategory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGalleryFilter_Active(t *testing.T) {
	tests := []struct {
		name string
		f    GalleryFilter
		want bool
	}{
		{"empty", GalleryFilter{}, false},
		{"types", GalleryFilter{Types: map[string]bool{"image": true}}, true},
		{"query", GalleryFilter{Query: "cat"}, true},
		{"exts", GalleryFilter{Exts: map[string]bool{".png": true}}, true},
		{"empty maps", GalleryFilter{Types: map[string]bool{}, Exts: map[string]bool{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Active(); got != tt.want {
				t.Errorf("Active() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGalleryFilter_Matches(t *testing.T) {
	img := GalleryItem{Filename: "Cat.PNG"}
	vid := GalleryItem{Filename: "clip.mp4"}
	dir := GalleryItem{Filename: "album", IsDir: true}

	tests := []struct {
		name string
		f    GalleryFilter
		gi   GalleryItem
		want bool
	}{
		{"empty matches all", GalleryFilter{}, img, true},
		{"type image matches image", GalleryFilter{Types: map[string]bool{"image": true}}, img, true},
		{"type image rejects video", GalleryFilter{Types: map[string]bool{"image": true}}, vid, false},
		{"type dir matches dir", GalleryFilter{Types: map[string]bool{"dir": true}}, dir, true},
		{"query case-insensitive substring", GalleryFilter{Query: "cat"}, img, true},
		{"query no match", GalleryFilter{Query: "dog"}, img, false},
		{"ext with dot matches (case-insensitive)", GalleryFilter{Exts: map[string]bool{".png": true}}, img, true},
		{"ext no match", GalleryFilter{Exts: map[string]bool{".jpg": true}}, img, false},
		{"AND: type and query both pass", GalleryFilter{Types: map[string]bool{"image": true}, Query: "cat"}, img, true},
		{"AND: type passes but query fails", GalleryFilter{Types: map[string]bool{"image": true}, Query: "dog"}, img, false},
		{"AND: type and ext conflict -> no match", GalleryFilter{Types: map[string]bool{"image": true}, Exts: map[string]bool{".mp4": true}}, img, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Matches(&tt.gi); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		{
			name:     "emoji file",
			path:     "/folder/",
			filename: "😀.jpg",
			isDir:    false,
			want:     "/fs/folder/%F0%9F%98%80.jpg",
		},
		{
			name:     "space file",
			path:     "/folder/",
			filename: "a b.png",
			isDir:    false,
			want:     "/fs/folder/a%20b.png",
		},
		{
			name:     "emoji folder",
			path:     "/folder/",
			filename: "😀",
			isDir:    true,
			want:     "?path=/folder/%F0%9F%98%80/",
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
		{
			name: "emoji path",
			path: "/hello/😀.jpg",
			want: "/fs/hello/%F0%9F%98%80.jpg",
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
		{
			name:     "emoji",
			path:     "/hellothere/",
			filename: "😀.jpg",
			want:     "/post/?path=/hellothere/%F0%9F%98%80.jpg",
		},
		{
			name:     "plus",
			path:     "/hellothere/",
			filename: "c++.png",
			want:     "/post/?path=/hellothere/c%2B%2B.png",
		},
		{
			name:     "space",
			path:     "/hellothere/",
			filename: "a b.png",
			want:     "/post/?path=/hellothere/a+b.png",
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
		{
			name:     "emoji",
			path:     "/hellothere/",
			filename: "😀.jpg",
			want:     "/resize/?path=/hellothere/%F0%9F%98%80.jpg",
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
		{
			name:     "emoji",
			path:     "/hellothere/",
			filename: "😀.mp4",
			want:     "/vidthumb/?path=/hellothere/%F0%9F%98%80.mp4",
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
