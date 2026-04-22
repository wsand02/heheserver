package cache

import (
	"image"

	"github.com/dgraph-io/ristretto/v2"
	ignore "github.com/wsand02/go-gitignore"
)

type IgnoreCache struct {
	*ristretto.Cache[string, *ignore.GitIgnore]
}

func NewIgnoreCache() (*IgnoreCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, *ignore.GitIgnore]{
		NumCounters: 1e4,     // 1000*10 seems ok for heheignore files...
		MaxCost:     1 << 24, // 16MB
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &IgnoreCache{cache}, nil
}

type ResizeCacheItem struct {
	image.Image
	Transparent bool
}

type ResizeCache struct {
	*ristretto.Cache[string, ResizeCacheItem]
}

func NewResizeCache() (*ResizeCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, ResizeCacheItem]{
		NumCounters: 1e7,     // 10M
		MaxCost:     1 << 30, // 1GB
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &ResizeCache{cache}, nil
}

type VidThumbCache struct {
	*ristretto.Cache[string, image.Image]
}

func NewVidThumbCache() (*VidThumbCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, image.Image]{
		NumCounters: 1e7,     // 10M
		MaxCost:     1 << 30, // 1GB
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &VidThumbCache{cache}, nil
}
