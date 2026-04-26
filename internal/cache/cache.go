package cache

import (
	"image"
	"log"

	"github.com/dgraph-io/ristretto/v2"
	ignore "github.com/wsand02/go-gitignore"
)

type IgnoreCache struct {
	*ristretto.Cache[string, *ignore.GitIgnore]
}

var ignoreCache *IgnoreCache

func sizeToNCMB(size int64) (int64, int64) {
	bytes := size * 1024 * 1024
	nc := size * 10000 // this hasnt been thought through thoroughly
	return bytes, nc
}

func GetIgnoreCache() *IgnoreCache {
	if ignoreCache == nil {
		log.Fatal("ignore cache not initialized")
	}
	return ignoreCache
}

var vidThumbCache *VidThumbCache

func GetVidThumbCache() *VidThumbCache {
	if vidThumbCache == nil {
		log.Fatal("vidthumb cache not initialized")
	}
	return vidThumbCache
}

var resizeCache *ResizeCache

func GetResizeCache() *ResizeCache {
	if resizeCache == nil {
		log.Fatal("resize cache not initialized")
	}
	return resizeCache
}

func NewIgnoreCache(size int64) error {
	size, nc := sizeToNCMB(size)
	cache, err := ristretto.NewCache(&ristretto.Config[string, *ignore.GitIgnore]{
		NumCounters: nc,   // 1000*10 seems ok for heheignore files...
		MaxCost:     size, // 16MB
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	ignoreCache = &IgnoreCache{cache}
	return nil
}

type ResizeCacheItem struct {
	image.Image
	Transparent bool
}

type ResizeCache struct {
	*ristretto.Cache[string, ResizeCacheItem]
}

func NewResizeCache(size int64) error {
	size, nc := sizeToNCMB(size)
	cache, err := ristretto.NewCache(&ristretto.Config[string, ResizeCacheItem]{
		NumCounters: nc,
		MaxCost:     size,
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	resizeCache = &ResizeCache{cache}
	return nil
}

type VidThumbCache struct {
	*ristretto.Cache[string, image.Image]
}

func NewVidThumbCache(size int64) error {
	size, nc := sizeToNCMB(size)
	cache, err := ristretto.NewCache(&ristretto.Config[string, image.Image]{
		NumCounters: nc,
		MaxCost:     size,
		BufferItems: 64,
	})
	if err != nil {
		return err
	}
	vidThumbCache = &VidThumbCache{cache}
	return nil
}
